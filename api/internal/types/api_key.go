package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// APIKey represents an API key for MCP authentication
type APIKey struct {
	bun.BaseModel  `bun:"table:api_keys,alias:ak" swaggerignore:"true"`
	ID             uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID         uuid.UUID  `json:"user_id" bun:"user_id,notnull,type:uuid"`
	OrganizationID uuid.UUID  `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	Name           string     `json:"name" bun:"name,notnull"`
	KeyHash        string     `json:"-" bun:"key_hash,notnull,unique"`
	Prefix         string     `json:"prefix" bun:"prefix,notnull"` // First few chars for display
	LastUsedAt     *time.Time `json:"last_used_at,omitempty" bun:"last_used_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" bun:"expires_at"`
	CreatedAt      time.Time  `json:"created_at" bun:"created_at,notnull,default:now()"`
	UpdatedAt      time.Time  `json:"updated_at" bun:"updated_at,notnull,default:now()"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty" bun:"revoked_at"`

	User         *User         `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Organization *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}

// IsValid checks if the API key is valid (not revoked and not expired)
func (ak *APIKey) IsValid() bool {
	if ak.RevokedAt != nil {
		return false
	}
	if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}
