package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	audit_storage "github.com/raghavyuva/nixopus-api/internal/features/audit/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type AuditService struct {
	storage *audit_storage.AuditStorage
	ctx     context.Context
	logger  logger.Logger
}

func NewAuditService(db *bun.DB, ctx context.Context, logger logger.Logger) *AuditService {
	return &AuditService{
		storage: audit_storage.NewAuditStorage(db, ctx),
		ctx:     ctx,
		logger:  logger,
	}
}

type AuditLogRequest struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Action         types.AuditAction
	ResourceType   types.AuditResourceType
	ResourceID     uuid.UUID
	OldValues      map[string]any
	NewValues      map[string]any
	Metadata       map[string]any
	IPAddress      string
	UserAgent      string
	RequestID      uuid.UUID
}

// ActivityMessage represents a human-readable activity
type ActivityMessage struct {
	ID          string                 `json:"id"`
	Message     string                 `json:"message"`
	Action      types.AuditAction      `json:"action"`
	Actor       string                 `json:"actor"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	Timestamp   string                 `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ActionColor string                 `json:"action_color"`
}

func (s *AuditService) LogAction(req *AuditLogRequest) error {
	auditLog := &types.AuditLog{
		ID:             uuid.New(),
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		Action:         req.Action,
		ResourceType:   req.ResourceType,
		ResourceID:     req.ResourceID,
		OldValues:      req.OldValues,
		NewValues:      req.NewValues,
		Metadata:       req.Metadata,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		CreatedAt:      time.Now(),
		RequestID:      req.RequestID,
	}

	if err := s.storage.CreateAuditLog(auditLog); err != nil {
		s.logger.Log(logger.Error, "Failed to create audit log", err.Error())
		return err
	}

	return nil
}

func (s *AuditService) GetAuditLogs(filters map[string]interface{}, page, pageSize int) ([]*types.AuditLog, int, error) {
	return s.storage.GetAuditLogs(filters, page, pageSize)
}

func (s *AuditService) GetAuditLogsByResource(resourceType types.AuditResourceType, resourceID uuid.UUID, page, pageSize int) ([]*types.AuditLog, int, error) {
	filters := map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
	}
	return s.storage.GetAuditLogs(filters, page, pageSize)
}

func (s *AuditService) GetAuditLogsByUser(userID uuid.UUID, page, pageSize int) ([]*types.AuditLog, int, error) {
	filters := map[string]interface{}{
		"user_id": userID,
	}
	return s.storage.GetAuditLogs(filters, page, pageSize)
}

func (s *AuditService) GetAuditLogsByOrganization(orgID uuid.UUID, page, pageSize int) ([]*types.AuditLog, int, error) {
	filters := map[string]interface{}{
		"organization_id": orgID,
	}
	return s.storage.GetAuditLogs(filters, page, pageSize)
}

// GetActivities converts audit logs to human-readable activities with filters
func (s *AuditService) GetActivities(filters map[string]interface{}, page, pageSize int) ([]*ActivityMessage, int, error) {
	auditLogs, totalCount, err := s.storage.GetAuditLogs(filters, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	activities := make([]*ActivityMessage, 0, len(auditLogs))
	for _, log := range auditLogs {
		activity := s.convertToActivity(log)
		if activity != nil {
			activities = append(activities, activity)
		}
	}

	return activities, totalCount, nil
}

// GetActivitiesByOrganization gets activities for a specific organization
func (s *AuditService) GetActivitiesByOrganization(orgID uuid.UUID, page, pageSize int, search string, resourceType string) ([]*ActivityMessage, int, error) {
	filters := map[string]interface{}{
		"organization_id": orgID,
	}

	if search != "" {
		filters["search"] = search
	}

	if resourceType != "" {
		filters["resource_type"] = resourceType
	}

	return s.GetActivities(filters, page, pageSize)
}

// convertToActivity converts an audit log to a human-readable activity message
func (s *AuditService) convertToActivity(log *types.AuditLog) *ActivityMessage {
	if log == nil {
		return nil
	}

	actor := "Unknown user"
	if log.User != nil && log.User.Username != "" {
		actor = log.User.Username
	} else if log.User != nil && log.User.Email != "" {
		actor = log.User.Email
	}

	// Generate the message based on resource type and action
	message := s.generateMessage(actor, log.Action, log.ResourceType, log.NewValues)

	actionColor := getActionColor(log.Action)

	return &ActivityMessage{
		ID:          log.ID.String(),
		Message:     message,
		Action:      log.Action,
		Actor:       actor,
		Resource:    string(log.ResourceType),
		ResourceID:  log.ResourceID.String(),
		Timestamp:   log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Metadata:    log.Metadata,
		ActionColor: actionColor,
	}
}

// generateMessage creates a human-readable message based on the audit log data
func (s *AuditService) generateMessage(actor string, action types.AuditAction, resourceType types.AuditResourceType, newValues map[string]any) string {
	switch resourceType {
	case types.AuditResourceUser:
		return s.generateUserMessage(actor, action, newValues)
	case types.AuditResourceOrganization:
		return s.generateOrganizationMessage(actor, action, newValues)
	case types.AuditResourceRole:
		return s.generateRoleMessage(actor, action, newValues)
	case types.AuditResourcePermission:
		return s.generatePermissionMessage(actor, action)
	case types.AuditResourceApplication:
		return s.generateApplicationMessage(actor, action, newValues)
	case types.AuditResourceDeployment:
		return s.generateDeploymentMessage(actor, action)
	case types.AuditResourceDomain:
		return s.generateDomainMessage(actor, action, newValues)
	case types.AuditResourceGithubConnector:
		return s.generateGithubConnectorMessage(actor, action)
	case types.AuditResourceSmtpConfig:
		return s.generateSmtpConfigMessage(actor, action)
	case types.AuditResourceNotification:
		return s.generateNotificationMessage(actor, action)
	case types.AuditResourceFeatureFlag:
		return s.generateFeatureFlagMessage(actor, action, newValues)
	case types.AuditResourceFileManager:
		return s.generateFileManagerMessage(actor, action, newValues)
	case types.AuditResourceContainer:
		return s.generateContainerMessage(actor, action)
	case types.AuditResourceAudit:
		return s.generateAuditMessage(actor, action)
	case types.AuditResourceTerminal:
		return s.generateTerminalMessage(actor, action)
	case types.AuditResourceIntegration:
		return s.generateIntegrationMessage(actor, action)
	default:
		return s.generateGenericMessage(actor, action, string(resourceType))
	}
}

// Individual message generators for each resource type
func (s *AuditService) generateUserMessage(actor string, action types.AuditAction, newValues map[string]any) string {
	switch action {
	case types.AuditActionCreate:
		if email, ok := newValues["email"].(string); ok && email != "" {
			return fmt.Sprintf("%s invited a new member %s", actor, email)
		}
		return fmt.Sprintf("%s added a new team member", actor)
	case types.AuditActionUpdate:
		if role, ok := newValues["role"].(string); ok && role != "" {
			return fmt.Sprintf("%s updated a member's role to %s", actor, role)
		}
		return fmt.Sprintf("%s updated a team member", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s removed a team member", actor)
	default:
		return fmt.Sprintf("%s %s a team member", actor, action)
	}
}

func (s *AuditService) generateOrganizationMessage(actor string, action types.AuditAction, newValues map[string]any) string {
	switch action {
	case types.AuditActionCreate:
		if name, ok := newValues["name"].(string); ok && name != "" {
			return fmt.Sprintf("%s created organization '%s'", actor, name)
		}
		return fmt.Sprintf("%s created an organization", actor)
	case types.AuditActionUpdate:
		if name, ok := newValues["name"].(string); ok && name != "" {
			return fmt.Sprintf("%s updated organization name to '%s'", actor, name)
		}
		if _, ok := newValues["description"].(string); ok {
			return fmt.Sprintf("%s updated organization description", actor)
		}
		return fmt.Sprintf("%s updated organization settings", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s deleted the organization", actor)
	default:
		return fmt.Sprintf("%s %s organization settings", actor, action)
	}
}

func (s *AuditService) generateApplicationMessage(actor string, action types.AuditAction, newValues map[string]any) string {
	switch action {
	case types.AuditActionCreate:
		if name, ok := newValues["name"].(string); ok && name != "" {
			return fmt.Sprintf("%s created application '%s'", actor, name)
		}
		return fmt.Sprintf("%s created a new application", actor)
	case types.AuditActionUpdate:
		if name, ok := newValues["name"].(string); ok && name != "" {
			return fmt.Sprintf("%s updated application '%s'", actor, name)
		}
		return fmt.Sprintf("%s updated an application", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s deleted an application", actor)
	default:
		return fmt.Sprintf("%s %s an application", actor, action)
	}
}

func (s *AuditService) generateDeploymentMessage(actor string, action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return fmt.Sprintf("%s triggered a new deployment", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated a deployment", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s cancelled a deployment", actor)
	default:
		return fmt.Sprintf("%s %s a deployment", actor, action)
	}
}

func (s *AuditService) generateFileManagerMessage(actor string, action types.AuditAction, newValues map[string]any) string {
	switch action {
	case types.AuditActionCreate:
		if fileName, ok := newValues["name"].(string); ok && fileName != "" {
			return fmt.Sprintf("%s created file '%s'", actor, fileName)
		}
		if path, ok := newValues["path"].(string); ok && path != "" {
			return fmt.Sprintf("%s created a file at %s", actor, path)
		}
		return fmt.Sprintf("%s created a file", actor)
	case types.AuditActionUpdate:
		if fileName, ok := newValues["name"].(string); ok && fileName != "" {
			return fmt.Sprintf("%s updated file '%s'", actor, fileName)
		}
		return fmt.Sprintf("%s updated a file", actor)
	case types.AuditActionDelete:
		if fileName, ok := newValues["name"].(string); ok && fileName != "" {
			return fmt.Sprintf("%s deleted file '%s'", actor, fileName)
		}
		return fmt.Sprintf("%s deleted a file", actor)
	default:
		return fmt.Sprintf("%s %s a file", actor, action)
	}
}

func (s *AuditService) generateDomainMessage(actor string, action types.AuditAction, newValues map[string]any) string {
	switch action {
	case types.AuditActionCreate:
		if domain, ok := newValues["domain"].(string); ok && domain != "" {
			return fmt.Sprintf("%s added domain '%s'", actor, domain)
		}
		return fmt.Sprintf("%s added a custom domain", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated domain settings", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s removed a domain", actor)
	default:
		return fmt.Sprintf("%s %s a domain", actor, action)
	}
}

func (s *AuditService) generateRoleMessage(actor string, action types.AuditAction, newValues map[string]any) string {
	switch action {
	case types.AuditActionCreate:
		if name, ok := newValues["name"].(string); ok && name != "" {
			return fmt.Sprintf("%s created role '%s'", actor, name)
		}
		return fmt.Sprintf("%s created a new role", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated role permissions", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s deleted a role", actor)
	default:
		return fmt.Sprintf("%s %s a role", actor, action)
	}
}

func (s *AuditService) generatePermissionMessage(actor string, action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return fmt.Sprintf("%s granted new permissions", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated permissions", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s revoked permissions", actor)
	default:
		return fmt.Sprintf("%s %s permissions", actor, action)
	}
}

func (s *AuditService) generateGithubConnectorMessage(actor string, action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return fmt.Sprintf("%s connected a GitHub repository", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated GitHub integration settings", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s disconnected GitHub integration", actor)
	default:
		return fmt.Sprintf("%s %s GitHub integration", actor, action)
	}
}

func (s *AuditService) generateSmtpConfigMessage(actor string, action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return fmt.Sprintf("%s configured SMTP settings", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated SMTP configuration", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s removed SMTP configuration", actor)
	default:
		return fmt.Sprintf("%s %s SMTP settings", actor, action)
	}
}

func (s *AuditService) generateNotificationMessage(actor string, action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return fmt.Sprintf("%s created a notification", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated notification settings", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s deleted a notification", actor)
	default:
		return fmt.Sprintf("%s %s notifications", actor, action)
	}
}

func (s *AuditService) generateFeatureFlagMessage(actor string, action types.AuditAction, newValues map[string]any) string {
	switch action {
	case types.AuditActionCreate:
		if name, ok := newValues["name"].(string); ok && name != "" {
			return fmt.Sprintf("%s enabled feature '%s'", actor, name)
		}
		return fmt.Sprintf("%s enabled a feature flag", actor)
	case types.AuditActionUpdate:
		if enabled, ok := newValues["enabled"].(bool); ok {
			if enabled {
				return fmt.Sprintf("%s enabled a feature flag", actor)
			} else {
				return fmt.Sprintf("%s disabled a feature flag", actor)
			}
		}
		return fmt.Sprintf("%s updated feature flag settings", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s removed a feature flag", actor)
	default:
		return fmt.Sprintf("%s %s feature flags", actor, action)
	}
}

func (s *AuditService) generateContainerMessage(actor string, action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return fmt.Sprintf("%s started a container", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated container configuration", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s stopped a container", actor)
	default:
		return fmt.Sprintf("%s %s a container", actor, action)
	}
}

func (s *AuditService) generateAuditMessage(actor string, action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return fmt.Sprintf("%s exported audit logs", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated audit settings", actor)
	default:
		return fmt.Sprintf("%s accessed audit logs", actor)
	}
}

func (s *AuditService) generateTerminalMessage(actor string, action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return fmt.Sprintf("%s opened a terminal session", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s executed commands in terminal", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s closed terminal session", actor)
	default:
		return fmt.Sprintf("%s %s terminal", actor, action)
	}
}

func (s *AuditService) generateIntegrationMessage(actor string, action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return fmt.Sprintf("%s added a new integration", actor)
	case types.AuditActionUpdate:
		return fmt.Sprintf("%s updated integration settings", actor)
	case types.AuditActionDelete:
		return fmt.Sprintf("%s removed an integration", actor)
	default:
		return fmt.Sprintf("%s %s an integration", actor, action)
	}
}

func (s *AuditService) generateGenericMessage(actor string, action types.AuditAction, resourceType string) string {
	return fmt.Sprintf("%s %s %s", actor, action, resourceType)
}

func getActionColor(action types.AuditAction) string {
	switch action {
	case types.AuditActionCreate:
		return "green"
	case types.AuditActionUpdate:
		return "blue"
	case types.AuditActionDelete:
		return "red"
	case types.AuditActionAccess:
		return "gray"
	default:
		return "gray"
	}
}
