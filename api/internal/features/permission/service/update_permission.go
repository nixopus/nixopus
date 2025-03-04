package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// UpdatePermission updates a permission in the application.
//
// It first checks if the permission exists using the provided ID.
// If the permission does not exist, it returns ErrPermissionDoesNotExist.
// If the permission exists, it updates the permission with the provided details and saves it to the database.
// If the update fails, it returns ErrFailedToUpdatePermission.
// Upon successful update, it returns nil.
func (c *PermissionService) UpdatePermission(permission *types.UpdatePermissionRequest) error {
	c.logger.Log(logger.Info, "Updating permission", "")
	existingPermission, err := c.storage.GetPermission(permission.ID)
	if err == nil && existingPermission.ID == uuid.Nil {
		c.logger.Log(logger.Error, types.ErrPermissionDoesNotExist.Error(), "")
		return types.ErrPermissionDoesNotExist
	}

	permissionToUpdate := shared_types.Permission{
		ID:          existingPermission.ID,
		Name:        permission.Name,
		Resource:    permission.Resource,
		Description: permission.Description,
		UpdatedAt:   time.Now(),
	}

	if err := c.storage.UpdatePermission(&permissionToUpdate); err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToUpdatePermission.Error(), err.Error())
		return types.ErrFailedToUpdatePermission
	}

	return nil
}
