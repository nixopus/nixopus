package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type AuthService struct {
	storage storage.AuthRepository
	Ctx     context.Context
	logger  logger.Logger
}

func NewAuthService(
	storage storage.AuthRepository,
	logger logger.Logger,
	ctx context.Context,
) *AuthService {
	return &AuthService{
		storage: storage,
		logger:  logger,
		Ctx:     ctx,
	}
}

// AuthServiceInterface is deprecated - Better Auth handles authentication
// Only GetUserByEmail is kept for backward compatibility
type AuthServiceInterface interface {
	GetUserByEmail(email string) (*shared_types.User, error)
}

func (s *AuthService) GetUserByEmail(email string) (*shared_types.User, error) {
	return s.storage.FindUserByEmail(email)
}
