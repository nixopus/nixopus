package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
)

func (c *RoleService) DeleteRole(id string) error {
	c.logger.Log(logger.Info, "Deleting role", "")
	existingRole, err := c.storage.GetRole(id)
	if err == nil && existingRole.ID == uuid.Nil {
		c.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), "")
		return types.ErrRoleDoesNotExist
	}

	if err := c.storage.DeleteRole(id); err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToDeleteRole.Error(), "")
		return types.ErrFailedToDeleteRole
	}

	return nil
}
