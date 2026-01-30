package auth

import (
	"context"

	auth_service "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type AuthController struct {
	validator    *validation.Validator
	service      auth_service.AuthServiceInterface
	store        *shared_storage.Store
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewAuthController(
	ctx context.Context,
	logger logger.Logger,
	notificationManager *notification.NotificationManager,
	authService auth_service.AuthService,
	store *shared_storage.Store,
) *AuthController {
	return &AuthController{
		validator:    validation.NewValidator(),
		service:      &authService,
		store:        store,
		ctx:          ctx,
		logger:       logger,
		notification: notificationManager,
	}
}
