package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// SSHKey represents an SSH key in the database
type SSHKey struct {
	bun.BaseModel       `bun:"table:ssh_keys,alias:sk" swaggerignore:"true"`
	ID                  uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID      uuid.UUID  `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	Name                string     `json:"name" bun:"name,notnull,type:varchar(255)"`
	Description         *string    `json:"description,omitempty" bun:"description,type:text"`
	Host                *string    `json:"host,omitempty" bun:"host,type:varchar(255)"`
	User                *string    `json:"user,omitempty" bun:"user,type:varchar(255)"`
	Port                *int       `json:"port,omitempty" bun:"port,type:integer,default:22"`
	PublicKey           string     `json:"public_key" bun:"public_key,notnull,type:text"`
	PrivateKeyEncrypted *string    `json:"private_key_encrypted,omitempty" bun:"private_key_encrypted,type:text"`
	PasswordEncrypted   *string    `json:"password_encrypted,omitempty" bun:"password_encrypted,type:text"`
	KeyType             string     `json:"key_type" bun:"key_type,type:varchar(20),default:'rsa'"`
	KeySize             int        `json:"key_size" bun:"key_size,type:integer,default:4096"`
	Fingerprint         *string    `json:"fingerprint,omitempty" bun:"fingerprint,type:varchar(255)"`
	AuthMethod          string     `json:"auth_method" bun:"auth_method,type:varchar(20),notnull,default:'key'"`
	IsActive            bool       `json:"is_active" bun:"is_active,notnull,default:true"`
	LastUsedAt          *time.Time `json:"last_used_at,omitempty" bun:"last_used_at,type:timestamp with time zone"`
	CreatedAt           time.Time  `json:"created_at" bun:"created_at,notnull,type:timestamp with time zone,default:now()"`
	UpdatedAt           time.Time  `json:"updated_at" bun:"updated_at,notnull,type:timestamp with time zone,default:now()"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" bun:"deleted_at,type:timestamp with time zone"`

	Organization *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}
