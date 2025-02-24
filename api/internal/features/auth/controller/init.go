package auth

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type AuthController struct {
	store     *shared_storage.Store
	validator *validation.Validator
	service   *service.AuthService
	ctx       context.Context
	logger    logger.Logger
}

// NewAuthController creates a new AuthController with the given App.
//
// This function creates a new AuthController with the given App and returns a pointer to it.
//
// The App passed to this function should be a valid App that has been created with storage.NewApp.
func NewAuthController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) *AuthController {
	return &AuthController{
		store:     store,
		validator: validation.NewValidator(),
		service:   service.NewAuthService(store, ctx, l),
		ctx:       ctx,
		logger:    l,
	}
}
