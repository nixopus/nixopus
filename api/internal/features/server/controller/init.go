package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/server/service"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type ServerController struct {
	store    *shared_storage.Store
	service  *service.ServerService
	ctx      context.Context
	logger   logger.Logger
	notifier shared_types.Notifier
}

func NewServerController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notifier shared_types.Notifier,
) *ServerController {
	return &ServerController{
		store:    store,
		service:  service.NewServerService(store, ctx, l),
		ctx:      ctx,
		logger:   l,
		notifier: notifier,
	}
}
