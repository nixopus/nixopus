package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type GithubConnector struct {
	bun.BaseModel `bun:"table:github_connectors,alias:gc"`

	ID             uuid.UUID  `bun:"id,pk,type:uuid" json:"id"`
	AppID          string     `bun:"app_id,notnull" json:"app_id"`
	Slug           string     `bun:"slug,notnull" json:"slug"`
	Pem            string     `bun:"pem,notnull" json:"pem"`
	ClientID       string     `bun:"client_id,notnull" json:"client_id"`
	ClientSecret   string     `bun:"client_secret,notnull" json:"client_secret"`
	WebhookSecret  string     `bun:"webhook_secret,notnull" json:"webhook_secret"`
	InstallationID string     `bun:"installation_id,notnull" json:"installation_id"`
	CreatedAt      time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt      *time.Time `bun:"deleted_at" json:"deleted_at"`
	UserID         uuid.UUID  `json:"user_id" bun:"user_id,notnull,type:uuid"`
}
