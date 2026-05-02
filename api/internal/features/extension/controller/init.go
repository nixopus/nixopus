package controller

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/cache"
	"github.com/nixopus/nixopus/api/internal/features/extension/service"
	"github.com/nixopus/nixopus/api/internal/features/extension/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
)

type ExtensionsController struct {
	store   *shared_storage.Store
	service *service.ExtensionService
	ctx     context.Context
	logger  logger.Logger
}

func NewExtensionsController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	appCache *cache.Cache,
) *ExtensionsController {
	storage := storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
	return &ExtensionsController{
		store:   store,
		service: service.NewExtensionService(ctx, l, &storage, appCache),
		ctx:     ctx,
		logger:  l,
	}
}
