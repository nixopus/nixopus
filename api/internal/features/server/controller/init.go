package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/server/service"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type ServerController struct {
	store        *shared_storage.Store
	service      *service.ServerService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewServerController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *ServerController {
	return &ServerController{
		store:        store,
		service:      service.NewServerService(store, ctx, l),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
	}
}
