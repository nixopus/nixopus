package auth

import (
	"context"

	auth_service "github.com/nixopus/nixopus/api/internal/features/auth/service"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

type AuthController struct {
	service *auth_service.AuthService
	ctx     context.Context
	logger  logger.Logger
}

func NewAuthController(
	ctx context.Context,
	l logger.Logger,
	service *auth_service.AuthService,
) *AuthController {
	return &AuthController{
		service: service,
		ctx:     ctx,
		logger:  l,
	}
}
