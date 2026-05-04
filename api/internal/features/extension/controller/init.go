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
	var svc *service.ExtensionService
	if appCache == nil {
		svc = service.NewExtensionService(ctx, l, &storage, nil)
	} else {
		svc = service.NewExtensionService(ctx, l, &storage, appCache)
	}
	return &ExtensionsController{
		store:   store,
		service: svc,
		ctx:     ctx,
		logger:  l,
	}
}
