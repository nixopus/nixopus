package auth

import (
	"context"

	auth_service "github.com/nixopus/nixopus/api/internal/features/auth/service"
	"github.com/nixopus/nixopus/api/internal/features/auth/validation"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type AuthController struct {
	validator *validation.Validator
	service   auth_service.AuthServiceInterface
	store     *shared_storage.Store
	ctx       context.Context
	logger    logger.Logger
	notifier  shared_types.Notifier
	cache     *auth_service.AuthCache
}

func NewAuthController(
	ctx context.Context,
	logger logger.Logger,
	notifier shared_types.Notifier,
	authService auth_service.AuthService,
	store *shared_storage.Store,
) *AuthController {
	return &AuthController{
		validator: validation.NewValidator(),
		service:   &authService,
		store:     store,
		ctx:       ctx,
		logger:    logger,
		notifier:  notifier,
		cache:     authService.Cache,
	}
}
