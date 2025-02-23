package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *PermissionService) CreatePermission(permission *types.CreatePermissionRequest) error {
	existingPermission, err := c.storage.GetPermissionByName(permission.Name)
	if err == nil && existingPermission.ID != uuid.Nil {
		return types.ErrPermissionAlreadyExists
	}

	permissionToCreate := shared_types.Permission{
		ID:          uuid.New(),
		Name:        permission.Name,
		Resource:    permission.Resource,
		Description: permission.Description,
	}

	if err := c.storage.CreatePermission(permissionToCreate); err != nil {
		return types.ErrFailedToCreatePermission
	}

	return nil
}
