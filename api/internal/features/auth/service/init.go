package service

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/auth/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type AuthService struct {
	storage storage.AuthRepository
	Cache   *AuthCache
	Ctx     context.Context
	logger  logger.Logger
}

func NewAuthService(
	storage storage.AuthRepository,
	l logger.Logger,
	ctx context.Context,
	redisURL string,
) *AuthService {
	var authCache *AuthCache
	if redisURL != "" {
		c, err := NewAuthCache(redisURL)
		if err != nil {
			l.Log(logger.Error, "failed to create auth cache, proceeding without cache", err.Error())
		} else {
			authCache = c
		}
	}

	return &AuthService{
		storage: storage,
		Cache:   authCache,
		logger:  l,
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
