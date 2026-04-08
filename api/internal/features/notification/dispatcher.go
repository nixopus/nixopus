package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification/channel"
	"github.com/nixopus/nixopus/api/internal/features/notification/helpers/preferences"
	"github.com/nixopus/nixopus/api/internal/features/notification/tasks"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/uptrace/bun"
)

// Dispatcher is the central notification coordinator. It replaces the old
// NotificationManager. Services call Emit() with a NotificationEvent;
// the Dispatcher checks preferences, resolves channels, and enqueues
// delivery tasks to the Redis-backed queue.
//
// Dispatcher implements shared_types.Notifier.
type Dispatcher struct {
	db          *bun.DB
	ctx         context.Context
	logger      logger.Logger
	prefManager *preferences.PreferenceManager
	channels    map[string]channel.Channel
}

// Compile-time check that Dispatcher implements Notifier.
var _ shared_types.Notifier = (*Dispatcher)(nil)

// NewDispatcher creates a new Dispatcher with the given channel adapters.
func NewDispatcher(db *bun.DB, ctx context.Context, l logger.Logger, channels map[string]channel.Channel) *Dispatcher {
	return &Dispatcher{
		db:          db,
		ctx:         ctx,
		logger:      l,
		prefManager: preferences.NewPreferenceManager(db, ctx),
		channels:    channels,
	}
}

// SetupQueue registers the notification delivery queue with taskq.
// Must be called after queue.Init() and before queue.StartConsumers().
func (d *Dispatcher) SetupQueue() {
	tasks.SetupNotificationQueue(d.channels, d.logger)
}

// Emit processes a notification event: checks user preferences,
// resolves which channels to use, and enqueues delivery tasks.
func (d *Dispatcher) Emit(event shared_types.NotificationEvent) error {
	if !skipPreferenceCheck[event.Type] {
		pref, ok := eventPreferenceMap[event.Type]
		if ok {
			shouldSend, err := d.prefManager.CheckUserNotificationPreferences(event.UserID, pref.Category, pref.Type)
			if err != nil {
				d.logger.Log(logger.Error, fmt.Sprintf("preference check failed for event %s: %s", event.Type, err.Error()), event.UserID)
			}
			if !shouldSend {
				d.logger.Log(logger.Info, fmt.Sprintf("notification suppressed by preference for event %s", event.Type), event.UserID)
				return nil
			}
		}
	}

	targetChannels := event.Channels
	if len(targetChannels) == 0 {
		targetChannels = defaultChannelsForEvent(event.Type)
	}

	userEmail, err := d.resolveUserEmail(event.UserID)
	if err != nil {
		d.logger.Log(logger.Error, fmt.Sprintf("failed to resolve user email: %s", err.Error()), event.UserID)
		userEmail = ""
	}

	var firstErr error
	for _, chName := range targetChannels {
		if _, ok := d.channels[chName]; !ok {
			d.logger.Log(logger.Error, fmt.Sprintf("channel %s not registered, skipping", chName), "")
			continue
		}

		if chName == "agent" && !d.isAgentEnabledForOrg(event.OrganizationID) {
			d.logger.Log(logger.Info, fmt.Sprintf("agent channel disabled for org %s, skipping", event.OrganizationID), event.UserID)
			continue
		}

		msg := d.buildMessage(event, chName, userEmail)

		if err := tasks.Enqueue(channel.DeliveryPayload{
			Channel:        chName,
			OrganizationID: event.OrganizationID,
			UserID:         event.UserID,
			Message:        msg,
		}); err != nil {
			d.logger.Log(logger.Error, fmt.Sprintf("failed to enqueue %s notification: %s", chName, err.Error()), event.UserID)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// SendDirect sends a notification immediately without preference checks
// or queue. Used by the /send API endpoint for explicit user-initiated sends.
func (d *Dispatcher) SendDirect(req SendNotificationRequest, userID string, organizationID string) SendNotificationResponse {
	ch, ok := d.channels[req.Channel]
	if !ok {
		return SendNotificationResponse{
			Channel: req.Channel,
			Success: false,
			Error:   fmt.Sprintf("unsupported channel: %s", req.Channel),
		}
	}

	recipient := req.To
	if recipient == "" && req.Channel == "email" {
		email, err := d.resolveUserEmail(userID)
		if err != nil {
			return SendNotificationResponse{
				Channel: req.Channel,
				Success: false,
				Error:   fmt.Sprintf("failed to resolve recipient: %s", err.Error()),
			}
		}
		recipient = email
	}

	subject := req.Subject
	if subject == "" {
		subject = "Notification from Nixopus"
	}

	msg := channel.Message{
		To:       recipient,
		Subject:  subject,
		Body:     req.Message,
		Metadata: map[string]string{"organization_id": organizationID},
	}

	if err := ch.Send(d.ctx, msg); err != nil {
		return SendNotificationResponse{
			Channel: req.Channel,
			Success: false,
			Error:   err.Error(),
		}
	}

	return SendNotificationResponse{Channel: req.Channel, Success: true}
}

func (d *Dispatcher) buildMessage(event shared_types.NotificationEvent, chName string, userEmail string) channel.Message {
	msg := channel.Message{
		Metadata: map[string]string{
			"organization_id": event.OrganizationID,
			"event_type":      string(event.Type),
			"user_id":         event.UserID,
		},
	}

	if chName == "email" {
		msg.To = userEmail
		if tmpl, ok := eventTemplateMap[event.Type]; ok {
			msg.Subject = tmpl.Subject
			msg.TemplateName = tmpl.Template
			msg.TemplateData = event.Data
		} else {
			msg.Subject = "Notification from Nixopus"
			msg.Body = fmt.Sprintf("Event: %s", event.Type)
		}
	} else if chName == "system_email" {
		msg.To = userEmail
		msg.TemplateData = event.Data
		if templateID, ok := systemEmailTemplateMap[event.Type]; ok {
			msg.Metadata["resend_template_id"] = templateID
		}
	} else {
		msg.Body = d.buildPlainTextBody(event)
	}

	if chName == "agent" {
		for key, val := range event.Data {
			if s, ok := val.(string); ok && s != "" {
				msg.Metadata[key] = s
			}
		}
	}

	return msg
}

func (d *Dispatcher) buildPlainTextBody(event shared_types.NotificationEvent) string {
	switch event.Type {
	case shared_types.EventUserAddedToOrg:
		return fmt.Sprintf("New user %s (%s) added to organization %s",
			getDataStr(event.Data, "UserName"),
			getDataStr(event.Data, "UserEmail"),
			getDataStr(event.Data, "OrganizationName"))
	case shared_types.EventUserRemovedFromOrg:
		return fmt.Sprintf("User %s (%s) removed from organization %s",
			getDataStr(event.Data, "UserName"),
			getDataStr(event.Data, "UserEmail"),
			getDataStr(event.Data, "OrganizationName"))
	case shared_types.EventDeploySuccess:
		return fmt.Sprintf("Deployment succeeded for %s", getDataStr(event.Data, "app_name"))
	case shared_types.EventDeployFailed:
		return fmt.Sprintf("Deployment failed for %s", getDataStr(event.Data, "app_name"))
	case shared_types.EventBuildFailed:
		return fmt.Sprintf("Build failed for %s: %s", getDataStr(event.Data, "app_name"), getDataStr(event.Data, "error_message"))
	case shared_types.EventHealthCheckCritical:
		return fmt.Sprintf("Health check critical for app %s endpoint %s (%s consecutive failures)",
			getDataStr(event.Data, "app_id"), getDataStr(event.Data, "endpoint"), getDataStr(event.Data, "consecutive_fails"))
	default:
		return fmt.Sprintf("Notification: %s", event.Type)
	}
}

func (d *Dispatcher) resolveUserEmail(userID string) (string, error) {
	var user shared_types.User
	err := d.db.NewSelect().
		Model(&user).
		Column("email").
		Where("id = ?", userID).
		Scan(d.ctx)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}
	return user.Email, nil
}

func (d *Dispatcher) isAgentEnabledForOrg(orgIDStr string) bool {
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return false
	}
	settings, err := utils.GetOrganizationSettings(d.ctx, d.db, orgID)
	if err != nil {
		return false
	}
	return settings.AIIncidentsEnabled != nil && *settings.AIIncidentsEnabled
}

func getDataStr(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
