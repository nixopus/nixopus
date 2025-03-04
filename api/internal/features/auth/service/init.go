package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	organization_service "github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	permissions_service "github.com/raghavyuva/nixopus-api/internal/features/permission/service"
	role_service "github.com/raghavyuva/nixopus-api/internal/features/role/service"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"

	organization_storage "github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	permission_storage "github.com/raghavyuva/nixopus-api/internal/features/permission/storage"
	roles_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
)

type AuthService struct {
	storage              storage.AuthRepository
	Ctx                  context.Context
	store                *shared_storage.Store
	logger               logger.Logger
	permissions_service  *permissions_service.PermissionService
	role_service         *role_service.RoleService
	organization_service *organization_service.OrganizationService
}

func NewAuthService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, auth_repo storage.AuthRepository, permissionRepository permission_storage.PermissionRepository, roleRepository roles_storage.RoleRepository, organizationRepository organization_storage.OrganizationRepository) *AuthService {
	return &AuthService{
		storage: auth_repo,
		store:   store,
		Ctx:     ctx,
		logger:  logger,
		permissions_service: permissions_service.NewPermissionService(
			store,
			ctx,
			logger,
			permissionRepository,
		),
		role_service: role_service.NewRoleService(
			store,
			ctx,
			logger,
			roleRepository,
		),
		organization_service: organization_service.NewOrganizationService(
			store,
			ctx,
			logger,
			organizationRepository,
		),
	}
}
