package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/domain/service"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type DomainsController struct {
	store        *shared_storage.Store
	service      *service.DomainsService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewDomainsController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *DomainsController {
	storage := storage.DomainStorage{DB: store.DB, Ctx: ctx}
	return &DomainsController{
		store:        store,
		service:      service.NewDomainsService(store, ctx, l, &storage),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
	}
}
