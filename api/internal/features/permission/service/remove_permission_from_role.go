package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
)

func (p *PermissionService) RemovePermissionFromRole(permissionID string, roleId string) error {
	role_storage := role_storage.RoleStorage{
		DB:  p.store.DB,
		Ctx: p.ctx,
	}
	existingRole, err := role_storage.GetRole(roleId)
	if err == nil && existingRole.ID == uuid.Nil {
		return types.ErrRoleDoesNotExist
	}

	existingPermission, err := p.storage.GetPermission(permissionID)
	if err == nil && existingPermission.ID == uuid.Nil {
		return types.ErrPermissionDoesNotExist
	}

	if err := p.storage.RemovePermissionFromRole(permissionID); err != nil {
		return types.ErrFailedToRemovePermissionFromRole
	}

	return nil
}
