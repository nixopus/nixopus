package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (p *PermissionService) GetPermissionByRole(roleID string) ([]shared_types.RolePermissions, error) {
	p.logger.Log(logger.Info, "Getting permissions by role", "")
	role_storage := role_storage.RoleStorage{
		DB:  p.store.DB,
		Ctx: p.ctx,
	}
	existingRole, err := role_storage.GetRole(roleID)
	if err == nil && existingRole.ID == uuid.Nil {
		p.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), "")
		return nil, types.ErrRoleDoesNotExist
	}

	permissions, err := p.storage.GetPermissionsByRole(roleID)
	if err != nil {
		p.logger.Log(logger.Error, types.ErrFailedToGetPermissionsByRole.Error(), err.Error())
		return nil, types.ErrFailedToGetPermissionsByRole
	}

	return permissions, nil
}
