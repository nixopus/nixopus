package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/service"
	"github.com/raghavyuva/nixopus-api/internal/features/user/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type UserController struct {
	store     *shared_storage.Store
	validator *validation.Validator
	service   *service.UserService
	ctx       context.Context
	logger    logger.Logger
}

func NewUserController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) *UserController {
	return &UserController{
		store:     store,
		validator: validation.NewValidator(),
		service:   service.NewUserService(store, ctx, l),
		ctx:       ctx,
		logger:    l,
	}
}
