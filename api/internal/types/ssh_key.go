package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// SSHKey represents an SSH key configuration for an organization
type SSHKey struct {
	bun.BaseModel `bun:"table:ssh_keys,alias:sk" swaggerignore:"true"`

	ID                  uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID      uuid.UUID  `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	Name                string     `json:"name" bun:"name,notnull"`
	Description         *string    `json:"description,omitempty" bun:"description"`
	Host                *string    `json:"host,omitempty" bun:"host"`
	User                *string    `json:"user,omitempty" bun:"user"`
	Port                *int       `json:"port,omitempty" bun:"port,default:22"`
	PublicKey           *string    `json:"public_key,omitempty" bun:"public_key"`
	PrivateKeyEncrypted *string    `json:"-" bun:"private_key_encrypted"`
	PasswordEncrypted   *string    `json:"-" bun:"password_encrypted"`
	KeyType             *string    `json:"key_type,omitempty" bun:"key_type,default:'rsa'"`
	KeySize             *int       `json:"key_size,omitempty" bun:"key_size,default:4096"`
	Fingerprint         *string    `json:"fingerprint,omitempty" bun:"fingerprint"`
	AuthMethod          string     `json:"auth_method" bun:"auth_method,notnull,default:'key'"`
	IsActive            bool       `json:"is_active" bun:"is_active,notnull,default:true"`
	LastUsedAt          *time.Time `json:"last_used_at,omitempty" bun:"last_used_at"`
	CreatedAt           time.Time  `json:"created_at" bun:"created_at,notnull,default:now()"`
	UpdatedAt           time.Time  `json:"updated_at" bun:"updated_at,notnull,default:now()"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Organization *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}
