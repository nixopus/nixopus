package types

import "errors"

type CreatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
}

type UpdatePermissionRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
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
	ErrRoleIDRequired                   = errors.New("role id is required to get permissions by role")
	ErrInvalidRequestType               = errors.New("invalid request type")
	ErrPermissionAlreadyExists          = errors.New("permission already exists")
	ErrPermissionDoesNotExist           = errors.New("permission does not exist")
	ErrRoleDoesNotExist                 = errors.New("role does not exist")
)

type AddPermissionToRoleRequest struct {
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
}

type RemovePermissionFromRoleRequest struct {
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
}

type GetPermissionRequest struct {
	ID string `json:"id"`
}
