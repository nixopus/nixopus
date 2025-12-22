package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	organization_service "github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type AuthService struct {
	storage              storage.AuthRepository
	Ctx                  context.Context
	logger               logger.Logger
	organization_service *organization_service.OrganizationService
}

func NewAuthService(
	storage storage.AuthRepository,
	logger logger.Logger,
	orgService *organization_service.OrganizationService,
	ctx context.Context,
) *AuthService {
	return &AuthService{
		storage:              storage,
		logger:               logger,
		Ctx:                  ctx,
		organization_service: orgService,
	}
}

type AuthServiceInterface interface {
	GenerateVerificationToken(userID string) (string, error)
	VerifyToken(token string) (string, error)
	MarkEmailAsVerified(userID string) error
	GetUserByID(userID string) (*shared_types.User, error)
	IsAdminRegistered() (bool, error)
	SetupTwoFactor(user *shared_types.User) (types.TwoFactorSetupResponse, error)
	VerifyTwoFactor(user *shared_types.User, code string) error
	DisableTwoFactor(user *shared_types.User) error
}
