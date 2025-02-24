package types

import "errors"

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

type GetRoleRequest struct {
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
	ErrInvalidRequestType = errors.New("invalid request type")
)
