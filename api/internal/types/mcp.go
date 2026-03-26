package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type MCPServer struct {
	bun.BaseModel `bun:"table:mcp_servers,alias:ms" swaggerignore:"true"`

	ID          uuid.UUID         `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	OrgID       uuid.UUID         `json:"org_id" bun:"org_id,notnull,type:uuid"`
	ProviderID  string            `json:"provider_id" bun:"provider_id,notnull"`
	Name        string            `json:"name" bun:"name,notnull"`
	Credentials map[string]string `json:"-" bun:"credentials,type:jsonb"`
	CustomURL   *string           `json:"custom_url,omitempty" bun:"custom_url"`
	Enabled     bool              `json:"enabled" bun:"enabled,notnull,default:true"`
	CreatedBy   uuid.UUID         `json:"created_by" bun:"created_by,notnull,type:uuid"`
	CreatedAt   time.Time         `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   time.Time         `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty" bun:"deleted_at"`
}
