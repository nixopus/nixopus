package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *PermissionService) UpdatePermission(permission *types.UpdatePermissionRequest) error {
	existingPermission, err := c.storage.GetPermission(permission.ID)
	if err == nil && existingPermission.ID == uuid.Nil {
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
		return  types.ErrFailedToUpdatePermission
	}

	return nil
}