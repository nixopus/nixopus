package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *PermissionService) AddPermissionToRole(permissionID string, roleID string) error {
	c.logger.Log(logger.Info, "Adding permission to role", "")
	role_storage := role_storage.RoleStorage{
		DB:  c.store.DB,
		Ctx: c.ctx,
	}
	existingRole, err := role_storage.GetRole(roleID)
	if err == nil && existingRole.ID == uuid.Nil {
		c.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), "")
		return types.ErrRoleDoesNotExist
	}

	existingPermission, err := c.storage.GetPermission(permissionID)
	if err == nil && existingPermission.ID == uuid.Nil {
		c.logger.Log(logger.Error, types.ErrPermissionDoesNotExist.Error(), "")
		return types.ErrPermissionDoesNotExist
	}

	rolePermissionToCreate := shared_types.RolePermissions{
		ID:           uuid.New(),
		RoleID:       existingRole.ID,
		PermissionID: existingPermission.ID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := c.storage.AddPermissionToRole(rolePermissionToCreate); err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToAddPermissionToRole.Error(), err.Error())
		return types.ErrFailedToAddPermissionToRole
	}

	return nil
}
