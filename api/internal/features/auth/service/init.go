package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	organization_service "github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	permissions_service "github.com/raghavyuva/nixopus-api/internal/features/permission/service"
	role_service "github.com/raghavyuva/nixopus-api/internal/features/role/service"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type AuthService struct {
	storage              storage.AuthRepository
	Ctx                  context.Context
	logger               logger.Logger
	permissions_service  *permissions_service.PermissionService
	role_service         *role_service.RoleService
	organization_service *organization_service.OrganizationService
}

func NewAuthService(
	storage storage.AuthRepository,
	logger logger.Logger,
	permissionService *permissions_service.PermissionService,
	roleService *role_service.RoleService,
	orgService *organization_service.OrganizationService,
	ctx context.Context,
) *AuthService {
	return &AuthService{
		storage:              storage,
		logger:               logger,
		Ctx:                  ctx,
		permissions_service:  permissionService,
		role_service:         roleService,
		organization_service: orgService,
	}
}

type AuthServiceInterface interface {
	Login(email string, password string) (types.AuthResponse, error)
	Logout(refreshToken string) error
	RefreshToken(refreshToken types.RefreshTokenRequest) (types.AuthResponse, error)
	Register(registrationRequest types.RegisterRequest) (types.AuthResponse, error)
	ResetPassword(user *shared_types.User, resetPasswordRequest types.ChangePasswordRequest) error
	GeneratePasswordResetLink(user *shared_types.User) error
}
