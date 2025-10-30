package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type ServersService struct {
	storage storage.ServerStorageInterface
	Ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewServersService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, server_repo storage.ServerStorageInterface) *ServersService {
	return &ServersService{
		storage: server_repo,
		store:   store,
		Ctx:     ctx,
		logger:  logger,
	}
}
