package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/domain/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type DomainsService struct {
	storage storage.DomainStorage
	Ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewDomainsService(store *shared_storage.Store, ctx context.Context, logger logger.Logger) *DomainsService {
	return &DomainsService{
		storage: storage.DomainStorage{
			DB:  store.DB,
			Ctx: ctx,
		},
		store:  store,
		Ctx:    ctx,
		logger: logger,
	}
}
