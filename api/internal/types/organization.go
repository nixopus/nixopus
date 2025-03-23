package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"table:roles,alias:r" swaggerignore:"true"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	Name          string     `json:"name" bun:"name,notnull,unique"`
	Description   string     `json:"description" bun:"description"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Permissions []Permission `json:"permissions,omitempty" bun:"m2m:role_permissions,join:Role=Permission"`
}

const (
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

type Permission struct {
	bun.BaseModel `bun:"table:permissions,alias:p" swaggerignore:"true"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	Name          string     `json:"name" bun:"name,notnull"`
	Description   string     `json:"description" bun:"description"`
	Resource      string     `json:"resource" bun:"resource,notnull"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Roles []Role `json:"roles,omitempty" bun:"m2m:role_permissions,join:Permission=Role"`
}

type RolePermissions struct {
	bun.BaseModel `bun:"table:role_permissions,alias:rp" swaggerignore:"true"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	RoleID        uuid.UUID  `json:"role_id" bun:"role_id,notnull,type:uuid"`
	PermissionID  uuid.UUID  `json:"permission_id" bun:"permission_id,notnull,type:uuid"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Role       *Role       `json:"role,omitempty" bun:"rel:belongs-to,join:role_id=id"`
	Permission *Permission `json:"permission,omitempty" bun:"rel:belongs-to,join:permission_id=id"`
}

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
	RoleID         uuid.UUID  `json:"role_id" bun:"role_id,notnull,type:uuid"`
	CreatedAt      time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Role         *Role         `json:"role,omitempty" bun:"rel:belongs-to,join:role_id=id"`
	User         *User         `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Organization *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
}

func (r *Role) NewRole(name string, description string) Role {
	return Role{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   nil,
	}
}
