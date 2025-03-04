package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetPermissionByRole retrieves the permissions associated with a specific role ID.
// It first verifies the existence of the role by its ID. If the role does not exist,
// it returns an error indicating the role does not exist. If the role exists, it
// fetches and returns the permissions linked to that role. If fetching permissions
// fails, it returns an error indicating the failure. Returns a slice of RolePermissions
// and an error, if any.
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

func (p *PermissionService) GetAllPermissions() ([]shared_types.Permission, error) {
	p.logger.Log(logger.Info, "Getting all permissions", "")
	permissions, err := p.storage.GetPermissions()
	if err != nil {
		p.logger.Log(logger.Error, types.ErrFailedToGetPermissions.Error(), err.Error())
		return nil, types.ErrFailedToGetPermissions
	}
	return permissions, nil
}
