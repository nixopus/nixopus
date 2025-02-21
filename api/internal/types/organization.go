package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"table:roles,alias:r"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	Name          string     `json:"name" bun:"name,notnull,unique"`
	Description   string     `json:"description" bun:"description"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Permissions []Permission `json:"permissions,omitempty" bun:"m2m:role_permissions,join:Role=Permission"`
}

type Permission struct {
	bun.BaseModel `bun:"table:permissions,alias:p"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	Name          string     `json:"name" bun:"name,notnull,unique"`
	Description   string     `json:"description" bun:"description"`
	Resource      string     `json:"resource" bun:"resource,notnull"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Roles []Role `json:"roles,omitempty" bun:"m2m:role_permissions,join:Permission=Role"`
}

type RolePermissions struct {
	bun.BaseModel `bun:"table:role_permissions,alias:rp"`
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
	bun.BaseModel `bun:"table:organizations,alias:o"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	Name          string     `json:"name" bun:"name,notnull,unique"`
	Description   string     `json:"description" bun:"description"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	Users []User `json:"users,omitempty" bun:"m2m:organization_users,join:Organization=User"`
}

type OrganizationUsers struct {
	bun.BaseModel  `bun:"table:organization_users,alias:ou"`
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

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name omitempty"`
	Description string `json:"description omitempty"`
}

type DeleteRoleRequest struct {
	ID string `json:"id"`
}

var (
	ErrRoleNameRequired   = errors.New("name is required to create a role")
	ErrFailedToCreateRole = errors.New("failed to create role")
	ErrFailedToGetRoles   = errors.New("failed to get roles")
	ErrFailedToGetRole    = errors.New("failed to get role")
	ErrRoleIDRequired     = errors.New("role id is required to get a role")
	ErrFailedToUpdateRole = errors.New("failed to update role")
	ErrRoleEmptyFields    = errors.New("name or description is required to update a role")
	ErrFailedToDeleteRole = errors.New("failed to delete role")
	ErrRoleAlreadyExists  = errors.New("role already exists")
	ErrRoleDoesNotExist   = errors.New("role does not exist")
)

type CreatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
}

type UpdatePermissionRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name omitempty"`
	Description string `json:"description omitempty"`
	Resource    string `json:"resource omitempty"`
}

type DeletePermissionRequest struct {
	ID string `json:"id"`
}

var (
	ErrPermissionNameRequired           = errors.New("name is required to create a permission")
	ErrFailedToCreatePermission         = errors.New("failed to create permission")
	ErrFailedToGetPermissions           = errors.New("failed to get permissions")
	ErrFailedToGetPermission            = errors.New("failed to get permission")
	ErrPermissionIDRequired             = errors.New("permission id is required to get a permission")
	ErrFailedToUpdatePermission         = errors.New("failed to update permission")
	ErrPermissionEmptyFields            = errors.New("name or description is required to update a permission")
	ErrFailedToDeletePermission         = errors.New("failed to delete permission")
	ErrPermissionResourceRequired       = errors.New("resource is required to create a permission")
	ErrFailedToAddPermissionToRole      = errors.New("failed to add permission to role")
	ErrFailedToRemovePermissionFromRole = errors.New("failed to remove permission from role")
	ErrFailedToGetPermissionsByRole     = errors.New("failed to get permissions by role")
)

type AddPermissionToRoleRequest struct {
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
}

type RemovePermissionFromRoleRequest struct {
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
}

var (
	ErrMissingOrganizationID              = errors.New("organization id is required to get organizations")
	ErrFailedToGetOrganizations           = errors.New("failed to get organizations")
	ErrFailedToGetOrganization            = errors.New("failed to get organization")
	ErrMissingOrganizationName            = errors.New("name is required to create an organization")
	ErrFailedToCreateOrganization         = errors.New("failed to create organization")
	ErrFailedToUpdateOrganization         = errors.New("failed to update organization")
	ErrFailedToDeleteOrganization         = errors.New("failed to delete organization")
	ErrFailedToGetOrganizationUsers       = errors.New("failed to get organization users")
	ErrFailedToAddUserToOrganization      = errors.New("failed to add user to organization")
	ErrOrganizationDoesNotExist           = errors.New("organization does not exist")
	ErrOrganizationAlreadyExists          = errors.New("organization already exists")
	ErrMissingUserID                      = errors.New("user id is required to add user to organization")
	ErrUserDoesNotExist                   = errors.New("user does not exist")
	ErrUserAlreadyInOrganization          = errors.New("user already in organization")
	ErrUserNotInOrganization              = errors.New("user not in organization")
	ErrFailedToRemoveUserFromOrganization = errors.New("failed to remove user from organization")
	ErrInternalServer                     = errors.New("internal server error")
	ErrInvalidRoleID                      = errors.New("invalid role id")
	ErrMissingRoleID                      = errors.New("role id is required to add user to organization")
	ErrOrganizationNameTooLong            = errors.New("organization name is too long")
	ErrOrganizationNameContainsSpaces     = errors.New("organization name contains spaces")
	ErrOrganizationDescriptionTooLong     = errors.New("organization description is too long")
	ErrInvalidOrganizationID              = errors.New("invalid organization id")
	ErrInvalidUserID                      = errors.New("invalid user id")
	ErrInvalidRequestType                 = errors.New("invalid request type")
)

type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateOrganizationRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name omitempty"`
	Description string `json:"description omitempty"`
}

type DeleteOrganizationRequest struct {
	ID string `json:"id"`
}

type AddUserToOrganizationRequest struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	RoleId         string `json:"role_id"`
}

type RemoveUserFromOrganizationRequest struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
}
