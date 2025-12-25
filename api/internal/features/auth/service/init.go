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
	Login(email string, password string) (types.AuthResponse, error)
	Logout(refreshToken string) error
	RefreshToken(refreshToken types.RefreshTokenRequest) (types.AuthResponse, error)
	Register(registrationRequest types.RegisterRequest, userType string) (types.AuthResponse, error)
	ResetPassword(user *shared_types.User, resetPasswordRequest types.ResetPasswordRequest) error
	GeneratePasswordResetLink(user *shared_types.User) (*shared_types.User, string, error)
	GetUserByResetToken(token string) (*shared_types.User, error)
	GenerateVerificationToken(userID string) (string, error)
	VerifyToken(token string) (string, error)
	MarkEmailAsVerified(userID string) error
	GetUserByID(userID string) (*shared_types.User, error)
	IsAdminRegistered() (bool, error)
	SetupTwoFactor(user *shared_types.User) (types.TwoFactorSetupResponse, error)
	VerifyTwoFactor(user *shared_types.User, code string) error
	DisableTwoFactor(user *shared_types.User) error
	VerifyTwoFactorCode(user *shared_types.User, code string) error
	GetUserByEmail(email string) (*shared_types.User, error)
}

func (s *AuthService) GetUserByEmail(email string) (*shared_types.User, error) {
	return s.storage.FindUserByEmail(email)
}
