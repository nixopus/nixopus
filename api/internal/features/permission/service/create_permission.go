package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreatePermission creates a new permission in the application.
//
// It first checks if a permission with the same name and resource already exists.
// If it does, it returns ErrPermissionAlreadyExists.
// If not, it creates a new permission with the provided details and saves it to the database.
// If the creation fails, it returns ErrFailedToCreatePermission.
// On successful creation, it returns nil.
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
