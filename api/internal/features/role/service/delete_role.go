package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
)

func (c *RoleService) DeleteRole(id string) error {
	existingRole, err := c.storage.GetRole(id)
	if err == nil && existingRole.ID == uuid.Nil {
		return  types.ErrRoleDoesNotExist
	}

	if err := c.storage.DeleteRole(id); err != nil {
		return types.ErrFailedToDeleteRole
	}

	return nil
}