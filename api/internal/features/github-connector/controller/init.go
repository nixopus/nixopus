package controller

import (
	"context"
	"net/http"

	"github.com/nixopus/nixopus/api/internal/features/github-connector/service"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/storage"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/validation"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type GithubConnectorController struct {
	store     *shared_storage.Store
	validator *validation.Validator
	service   *service.GithubConnectorService
	ctx       context.Context
	logger    logger.Logger
	notifier  shared_types.Notifier
}

func NewGithubConnectorController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notifier shared_types.Notifier,
) *GithubConnectorController {
	storage := storage.GithubConnectorStorage{DB: store.DB, Ctx: ctx}
	return &GithubConnectorController{
		store:     store,
		validator: validation.NewValidator(&storage),
		service:   service.NewGithubConnectorService(store, ctx, l, &storage),
		ctx:       ctx,
		logger:    l,
		notifier:  notifier,
	}
}

func (c *GithubConnectorController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	user := utils.GetUser(w, r)

	if user == nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToGetUserFromContext.Error(), shared_types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return false
	}

	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}
