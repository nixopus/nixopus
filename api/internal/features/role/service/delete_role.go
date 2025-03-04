package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
)

// DeleteRole deletes a role in the application.
//
// It first checks if the role exists using the provided ID.
// If the role does not exist, it returns ErrRoleDoesNotExist.
// If the role exists, it deletes the role from the database.
// If the deletion fails, it returns ErrFailedToDeleteRole.
// Upon successful deletion, it returns nil.
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
