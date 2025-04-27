package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type FeatureFlag struct {
	bun.BaseModel  `bun:"table:feature_flags,alias:ff" swaggerignore:"true"`
	ID             uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID uuid.UUID  `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	FeatureName    string     `json:"feature_name" bun:"feature_name,notnull"`
	IsEnabled      bool       `json:"is_enabled" bun:"is_enabled,notnull,default:true"`
	CreatedAt      time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Organization *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}

type FeatureName string

const (
	FeatureTerminal        FeatureName = "terminal"
	FeatureFileManager     FeatureName = "file_manager"
	FeatureMonitoring      FeatureName = "monitoring"
	FeatureProxyConfig     FeatureName = "proxy_config"
	FeatureGithubConnector FeatureName = "github_connector"
	FeatureAudit           FeatureName = "audit"
	FeatureNotifications   FeatureName = "notifications"
	FeatureDomain          FeatureName = "domain"
	FeatureSelfHosted      FeatureName = "self_hosted"
)

type UpdateFeatureFlagRequest struct {
	FeatureName string `json:"feature_name" validate:"required"`
	IsEnabled   bool   `json:"is_enabled"`
}

type GetFeatureFlagsResponse struct {
	Features []FeatureFlag `json:"features"`
}
