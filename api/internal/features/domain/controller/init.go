package controller

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/domain/service"
	"github.com/nixopus/nixopus/api/internal/features/domain/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type DomainsController struct {
	store    *shared_storage.Store
	service  *service.DomainsService
	ctx      context.Context
	logger   logger.Logger
	notifier shared_types.Notifier
}

func NewDomainsController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notifier shared_types.Notifier,
) *DomainsController {
	domainStorage := storage.DomainStorage{DB: store.DB, Ctx: ctx, Logger: &l}
	return &DomainsController{
		store:    store,
		service:  service.NewDomainsService(ctx, l, &domainStorage),
		ctx:      ctx,
		logger:   l,
		notifier: notifier,
	}
}
