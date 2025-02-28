package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *PermissionService) CreatePermission(permission *types.CreatePermissionRequest) error {
	c.logger.Log(logger.Info, "Creating permission", "")
	existingPermission, err := c.storage.GetPermissionByNameAndResource(permission.Name, permission.Resource)
	if err == nil && existingPermission.ID != uuid.Nil {
		c.logger.Log(logger.Error, types.ErrPermissionAlreadyExists.Error(), "")
		return types.ErrPermissionAlreadyExists
	}

	permissionToCreate := shared_types.Permission{
		ID:          uuid.New(),
		Name:        permission.Name,
		Resource:    permission.Resource,
		Description: permission.Description,
	}

	if err := c.storage.CreatePermission(permissionToCreate); err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreatePermission.Error(), err.Error())
		return types.ErrFailedToCreatePermission
	}

	return nil
}
