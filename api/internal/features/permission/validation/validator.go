package validation

import (
	"encoding/json"
	"io"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
)

type Validator struct {
	storage storage.PermissionRepository
}

func NewValidator(storage storage.PermissionRepository) *Validator {
	return &Validator{
		storage: storage,
	}
}

func validateCreatePermissionRequest(permission types.CreatePermissionRequest) error {
	if permission.Name == "" {
		return types.ErrPermissionNameRequired
	}
	if permission.Resource == "" {
		return types.ErrPermissionResourceRequired
	}
	return nil
}

func validateGetPermissionRequest(id string) error {
	if id == "" {
		return types.ErrPermissionIDRequired
	}
	return nil
}

func (v *Validator) validateUpdatePermissionRequest(permission types.UpdatePermissionRequest) error {
	if permission.Name == "" && permission.Description == "" {
		return types.ErrPermissionEmptyFields
	}

	if permission.ID == "" {
		return types.ErrPermissionIDRequired
	}

	existing_permission, err := v.storage.GetPermission(permission.ID)
	if err != nil {
		return err
	}

	if existing_permission.ID == uuid.Nil {
		return types.ErrPermissionDoesNotExist
	}

	return nil
}

func validateAddPermissionToRoleRequest(permission types.AddPermissionToRoleRequest) error {
	if permission.PermissionID == "" {
		return types.ErrPermissionIDRequired
	}
	if permission.RoleID == "" {
		return types.ErrRoleIDRequired
	}
	return nil
}

func validateRemovePermissionFromRoleRequest(permission types.RemovePermissionFromRoleRequest) error {
	if permission.PermissionID == "" {
		return types.ErrPermissionIDRequired
	}
	if permission.RoleID == "" {
		return types.ErrRoleIDRequired
	}
	return nil
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.CreatePermissionRequest:
		return validateCreatePermissionRequest(*r)
	case *types.GetPermissionRequest:
		return validateGetPermissionRequest(r.ID)
	case *types.DeletePermissionRequest:
		return validateGetPermissionRequest(r.ID)
	case *types.UpdatePermissionRequest:
		return v.validateUpdatePermissionRequest(*r)
	case *types.AddPermissionToRoleRequest:
		return validateAddPermissionToRoleRequest(*r)
	case *types.RemovePermissionFromRoleRequest:
		return validateRemovePermissionFromRoleRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}
