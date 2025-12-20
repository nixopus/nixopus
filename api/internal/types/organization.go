package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Role struct{}

const (
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

type Organization struct {
	bun.BaseModel `bun:"table:organizations,alias:o" swaggerignore:"true"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	Name          string     `json:"name" bun:"name,notnull,unique"`
	Description   string     `json:"description" bun:"description"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Users []User `json:"users,omitempty" bun:"m2m:organization_users,join:Organization=User"`
}

type OrganizationUsers struct {
	bun.BaseModel  `bun:"table:organization_users,alias:ou" swaggerignore:"true"`
	ID             uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	UserID         uuid.UUID  `json:"user_id" bun:"user_id,notnull,type:uuid"`
	OrganizationID uuid.UUID  `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	CreatedAt      time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	User         *User         `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Organization *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}

// OrganizationUsersWithRoles represents organization users with their roles and permissions from SuperTokens
type OrganizationUsersWithRoles struct {
	OrganizationUsers
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
