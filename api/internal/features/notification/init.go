package notification

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/notification/helpers/discord"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/helpers/email"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/helpers/preferences"
	slackhelper "github.com/raghavyuva/nixopus-api/internal/features/notification/helpers/slack"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

func NewNotificationPayload(payloadType NotificationPayloadType, userID string, data interface{}, category NotificationCategory) NotificationPayload {
	return NotificationPayload{
		Type:      payloadType,
		UserID:    userID,
		Timestamp: time.Now(),
		Data:      data,
		Category:  category,
	}
}

func NewNotificationManager(db *bun.DB) *NotificationManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &NotificationManager{
		db:             db,
		ctx:            ctx,
		cancel:         cancel,
		PayloadChan:    make(chan NotificationPayload, 100),
		prefManager:    preferences.NewPreferenceManager(db, ctx),
		emailManager:   email.NewEmailManager(db, ctx),
		slackManager:   slackhelper.NewSlackManager(),
		discordManager: discord.NewDiscordManager(),
	}
}

// Start starts the notification manager to listen for notifications from a go routine
// based on the type of the notification that we receive we can check the notification preferences of the user
// and then send the notification to the respective channel
func (m *NotificationManager) Start() {
	go func() {
		for {
			select {
			case payload := <-m.PayloadChan:
				switch payload.Category {
				case NotificationCategoryAuthentication:
					fmt.Printf("Authentication Notification - %+v", payload)
					if payload.Type == NotificationPayloadTypePasswordReset {
						m.SendPasswordResetNotification(payload)
					}
					if payload.Type == NotificationPayloadTypeVerificationEmail {
						m.SendVerificationEmailNotification(payload)
					}
					if payload.Type == NotificationPayloadTypeLogin {
						m.SendLoginNotification(payload)
					}
				case NotificationCategoryOrganization:
					m.SendOrganizationNotification(payload)
				}
			case <-m.ctx.Done():
				return
			}
		}
	}()
}

func (m *NotificationManager) Stop() {
	m.cancel()
}

// SendNotification sends a notification
func (m *NotificationManager) SendNotification(payload NotificationPayload) {
	m.PayloadChan <- payload
}

// SendLoginNotification sends a login notification to the user
func (m *NotificationManager) SendLoginNotification(payload NotificationPayload) {
	fmt.Printf("Login Notification - %+v", payload)
	if data, ok := payload.Data.(NotificationAuthenticationData); ok {
		shouldSend, err := m.prefManager.CheckUserNotificationPreferences(payload.UserID, string(NotificationCategoryAuthentication), "login-alerts")
		if err != nil {
			log.Printf("Failed to check notification preferences: %s", err)
		}
		fmt.Printf("Should send login notification: %t for user %s and type %s", shouldSend, payload.UserID, string(NotificationCategoryAuthentication))
		if shouldSend {
			err := m.emailManager.SendEmailWithTemplate(payload.UserID, email.EmailData{
				Subject:  "Login Notification",
				Template: "login_notification.html",
				Data: map[string]interface{}{
					"IP":       data.IP,
					"Browser":  data.Browser,
					"Email":    data.Email,
					"UserName": data.UserName,
				},
				Type:        "login-alerts",
				ContentType: "text/html; charset=UTF-8",
				Category:    string(shared_types.SecurityCategory),
			})
			if err != nil {
				log.Printf("Failed to send login notification email: %s", err)
			}
		}
	}
}

// SendPasswordResetNotification sends a password reset notification to the user
func (m *NotificationManager) SendPasswordResetNotification(payload NotificationPayload) {
	fmt.Printf("Password Reset Notification - %+v", payload)

	// we need not to check the notification preferences for password reset notifications
	if data, ok := payload.Data.(NotificationPasswordResetData); ok {
		m.emailManager.SendPasswordResetEmail(payload.UserID, data.Token)
	}
}

// SendVerificationEmailNotification sends a verification email notification to the user
func (m *NotificationManager) SendVerificationEmailNotification(payload NotificationPayload) {
	fmt.Printf("Verification Email Notification - %+v", payload)
	if data, ok := payload.Data.(NotificationVerificationEmailData); ok {
		m.emailManager.SendVerificationEmail(payload.UserID, data.Token)
	}
}

func (m *NotificationManager) sendWebhookNotification(userID string, message string) {
	webhookUrl, err := m.GetWebhookURL(userID, "slack")
	if err != nil {
		log.Printf("Failed to get slack webhook URL: %s", err)
	} else {
		m.slackManager.SendNotification(message, webhookUrl)
	}

	webhookURL, err := m.GetWebhookURL(userID, "discord")
	if err != nil {
		log.Printf("Failed to get discord webhook URL: %s", err)
	} else {
		m.discordManager.SendNotification(message, webhookURL)
	}
}

func (m *NotificationManager) SendOrganizationNotification(payload NotificationPayload) {
	shouldSend, err := m.prefManager.CheckUserNotificationPreferences(payload.UserID, string(ActivityCategory), "team-updates")
	if err != nil {
		log.Printf("Failed to check notification preferences: %s", err)
		return
	}

	if !shouldSend {
		return
	}

	switch payload.Type {
	case NotificationPayloadTypeAddUserToOrganization:
		if data, ok := payload.Data.(AddUserToOrganizationData); ok {
			err := m.emailManager.SendAddUserToOrganizationEmail(payload.UserID, email.AddUserToOrganizationData{
				OrganizationName: data.OrganizationName,
				UserName:         data.UserName,
				UserEmail:        data.UserEmail,
				IP:               data.IP,
				Browser:          data.Browser,
			})
			if err != nil {
				log.Printf("Failed to send add user to organization email: %s", err)
			}
			m.sendWebhookNotification(payload.UserID, fmt.Sprintf("New user %s (%s) added to organization %s", data.UserName, data.UserEmail, data.OrganizationName))
		}
	case NotificationPayloadTypeRemoveUserFromOrganization:
		if data, ok := payload.Data.(RemoveUserFromOrganizationData); ok {
			err := m.emailManager.SendRemoveUserFromOrganizationEmail(payload.UserID, email.RemoveUserFromOrganizationData{
				OrganizationName: data.OrganizationName,
				UserName:         data.UserName,
				UserEmail:        data.UserEmail,
				IP:               data.IP,
				Browser:          data.Browser,
			})
			if err != nil {
				log.Printf("Failed to send remove user from organization email: %s", err)
			}
			m.sendWebhookNotification(payload.UserID, fmt.Sprintf("User %s (%s) removed from organization %s", data.UserName, data.UserEmail, data.OrganizationName))
		}
	}
}

func (m *NotificationManager) GetWebhookURL(userID string, webhookType string) (string, error) {
	var config shared_types.WebhookConfig

	err := m.db.NewSelect().
		Model(&config).
		Where("type = ?", webhookType).
		Where("user_id = ?", userID).
		Where("is_active = ?", true).
		Scan(m.ctx)

	if err != nil {
		return "", err
	}

	return config.WebhookURL, nil
}
