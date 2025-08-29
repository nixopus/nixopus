package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
	AuditActionAccess AuditAction = "access"
)

type AuditResourceType string

const (
	AuditResourceUser            AuditResourceType = "user"
	AuditResourceOrganization    AuditResourceType = "organization"
	AuditResourceRole            AuditResourceType = "role"
	AuditResourcePermission      AuditResourceType = "permission"
	AuditResourceApplication     AuditResourceType = "application"
	AuditResourceDeployment      AuditResourceType = "deployment"
	AuditResourceDomain          AuditResourceType = "domain"
	AuditResourceGithubConnector AuditResourceType = "github_connector"
	AuditResourceSmtpConfig      AuditResourceType = "smtp_config"
	AuditResourceNotification    AuditResourceType = "notification"
	AuditResourceFeatureFlag     AuditResourceType = "feature_flag"
	AuditResourceFileManager     AuditResourceType = "file_manager"
	AuditResourceContainer       AuditResourceType = "container"
	AuditResourceAudit           AuditResourceType = "audit"
	AuditResourceTerminal        AuditResourceType = "terminal"
	AuditResourceIntegration     AuditResourceType = "integration"
)

type AuditLog struct {
	bun.BaseModel  `bun:"table:audit_logs,alias:al"`
	ID             uuid.UUID         `json:"id" bun:"id,pk,type:uuid"`
	UserID         uuid.UUID         `json:"user_id" bun:"user_id,type:uuid"`
	OrganizationID uuid.UUID         `json:"organization_id" bun:"organization_id,type:uuid"`
	Action         AuditAction       `json:"action" bun:"action,notnull"`
	ResourceType   AuditResourceType `json:"resource_type" bun:"resource_type,notnull"`
	ResourceID     uuid.UUID         `json:"resource_id" bun:"resource_id,notnull,type:uuid"`
	OldValues      map[string]any    `json:"old_values,omitempty" bun:"old_values,type:jsonb"`
	NewValues      map[string]any    `json:"new_values,omitempty" bun:"new_values,type:jsonb"`
	Metadata       map[string]any    `json:"metadata,omitempty" bun:"metadata,type:jsonb"`
	IPAddress      string            `json:"ip_address,omitempty" bun:"ip_address"`
	UserAgent      string            `json:"user_agent,omitempty" bun:"user_agent"`
	CreatedAt      time.Time         `json:"created_at" bun:"created_at,notnull"`
	RequestID      uuid.UUID         `json:"request_id,omitempty" bun:"request_id,type:uuid"`
	User           *User             `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Organization   *Organization     `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}
