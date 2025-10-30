package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Domain struct {
	bun.BaseModel  `bun:"table:domains,alias:do" swaggerignore:"true"`
	ID             uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	UserID         uuid.UUID  `json:"user_id" bun:"user_id,notnull,type:uuid"`
	CreatedAt      time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`
	Name           string     `json:"name" bun:"name,notnull"`
	OrganizationID uuid.UUID  `json:"organization_id" bun:"organization_id,notnull"`
}

type Server struct {
	bun.BaseModel     `bun:"table:servers,alias:s" swaggerignore:"true"`
	ID                uuid.UUID     `json:"id" bun:"id,pk,type:uuid"`
	Name              string        `json:"name" bun:"name,notnull"`
	Description       string        `json:"description" bun:"description"`
	Host              string        `json:"host" bun:"host,notnull"`
	Port              int           `json:"port" bun:"port,notnull"`
	Username          string        `json:"username" bun:"username,notnull"`
	SSHPassword       *string       `json:"ssh_password,omitempty" bun:"ssh_password"`
	SSHPrivateKeyPath *string       `json:"ssh_private_key_path,omitempty" bun:"ssh_private_key_path"`
	Status            string        `json:"status" bun:"status,notnull,default:'inactive'"`
	CreatedAt         time.Time     `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt         time.Time     `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt         *time.Time    `json:"deleted_at,omitempty" bun:"deleted_at"`
	UserID            uuid.UUID     `json:"user_id" bun:"user_id,notnull,type:uuid"`
	OrganizationID    uuid.UUID     `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	User              *User         `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Organization      *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}
