package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
)

func (p *PermissionService) RemovePermissionFromRole(permissionID string, roleId string) error {
	p.logger.Log(logger.Info, "Removing permission from role", "")
	role_storage := role_storage.RoleStorage{
		DB:  p.store.DB,
		Ctx: p.ctx,
	}
	existingRole, err := role_storage.GetRole(roleId)
	if err == nil && existingRole.ID == uuid.Nil {
		p.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), "")
		return types.ErrRoleDoesNotExist
	}

	existingPermission, err := p.storage.GetPermission(permissionID)
	if err == nil && existingPermission.ID == uuid.Nil {
		p.logger.Log(logger.Error, types.ErrPermissionDoesNotExist.Error(), "")
		return types.ErrPermissionDoesNotExist
	}

	if err := p.storage.RemovePermissionFromRole(permissionID); err != nil {
		p.logger.Log(logger.Error, types.ErrFailedToRemovePermissionFromRole.Error(), err.Error())
		return types.ErrFailedToRemovePermissionFromRole
	}

	return nil
}
