package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type NotificationPreferences struct {
	bun.BaseModel `bun:"table:notification_preferences,alias:np" swaggerignore:"true"`
	ID            uuid.UUID        `json:"id" bun:"id,pk,type:uuid"`
	UserID        uuid.UUID        `json:"user_id" bun:"user_id,notnull,type:uuid"`
	CreatedAt     time.Time        `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time        `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time       `json:"deleted_at,omitempty" bun:"deleted_at"`
	Items         []PreferenceItem `json:"items" bun:"rel:has-many,join:id=preference_id"`
}

type PreferenceItem struct {
	bun.BaseModel `bun:"table:preference_item,alias:npi" swaggerignore:"true"`
	ID            uuid.UUID `json:"id" bun:"id,pk,type:uuid"`
	PreferenceID  uuid.UUID `json:"preference_id" bun:"preference_id,notnull,type:uuid"`
	Category      string    `json:"category" bun:"category,notnull"`
	Type          string    `json:"type" bun:"type,notnull"`
	Enabled       bool      `json:"enabled" bun:"enabled,notnull"`
	CreatedAt     time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
}

type PreferenceCategory string

const (
	ActivityCategory PreferenceCategory = "activity"
	SecurityCategory PreferenceCategory = "security"
	UpdateCategory   PreferenceCategory = "update"
)

func CreateDefaultPreferenceItems(preferenceID uuid.UUID) []PreferenceItem {
	now := time.Now()
	items := []PreferenceItem{
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     string(ActivityCategory),
			Type:         "team-updates",
			Enabled:      true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     string(SecurityCategory),
			Type:         "login-alerts",
			Enabled:      true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     string(SecurityCategory),
			Type:         "password-changes",
			Enabled:      true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     string(SecurityCategory),
			Type:         "security-alerts",
			Enabled:      true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     string(UpdateCategory),
			Type:         "product-updates",
			Enabled:      true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     string(UpdateCategory),
			Type:         "newsletter",
			Enabled:      false,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     string(UpdateCategory),
			Type:         "marketing",
			Enabled:      false,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	return items
}

type SMTPConfigs struct {
	bun.BaseModel  `bun:"table:smtp_configs,alias:sc" swaggerignore:"true"`
	ID             uuid.UUID `json:"id" bun:"id,pk,type:uuid"`
	Host           string    `json:"host" bun:"host,notnull"`
	Port           int       `json:"port" bun:"port,notnull"`
	Username       string    `json:"username" bun:"username,notnull"`
	Password       string    `json:"-" bun:"password,notnull"`
	FromEmail      string    `json:"from_email" bun:"from_email,notnull"`
	FromName       string    `json:"from_name" bun:"from_name,notnull"`
	Security       string    `json:"security" bun:"security,notnull"`
	CreatedAt      time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	IsActive       bool      `json:"is_active" bun:"is_active,notnull,default:false"`
	UserID         uuid.UUID `json:"user_id" bson:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id" bun:"organization_id,notnull"`
}

type WebhookConfig struct {
	bun.BaseModel  `bun:"table:webhook_configs,alias:wc" swaggerignore:"true"`
	ID             uuid.UUID `json:"id" bun:"id,pk,type:uuid"`
	Type           string    `json:"type" bun:"type,notnull"`
	WebhookURL     string    `json:"webhook_url" bun:"webhook_url,notnull"`
	WebhookSecret  *string   `json:"webhook_secret,omitempty" bun:"webhook_secret"`
	ChannelID      string    `json:"channel_id" bun:"channel_id,notnull"`
	IsActive       bool      `json:"is_active" bun:"is_active,notnull,default:false"`
	UserID         uuid.UUID `json:"user_id" bun:"user_id,notnull,type:uuid"`
	OrganizationID uuid.UUID `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	CreatedAt      time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
}
