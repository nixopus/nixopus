package controller

import (
	"context"
	"errors"
	"net/http"

	cache "github.com/nixopus/nixopus/api/internal/cache"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/service"
	"github.com/nixopus/nixopus/api/internal/features/user/storage"
	"github.com/nixopus/nixopus/api/internal/features/user/validation"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type UserController struct {
	validator *validation.Validator
	service   *service.UserService
	ctx       context.Context
	logger    logger.Logger
	cache     *cache.Cache
}

func NewUserController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	cache *cache.Cache,
) *UserController {
	return &UserController{
		validator: validation.NewValidator(),
		service:   service.NewUserService(store, ctx, l, &storage.UserStorage{DB: store.DB, Ctx: ctx}),
		ctx:       ctx,
		logger:    l,
		cache:     cache,
	}
}

func (c *UserController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) error {
	user := utils.GetUser(w, r)

	if user == nil {
		return errors.New("authentication required")
	}

	if err := c.validator.ValidateRequest(req, *user); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		return err
	}

	return nil
}
