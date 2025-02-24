package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type PermissionService struct {
	storage storage.PermissionStorage
	ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewPermissionService(store *shared_storage.Store, ctx context.Context, logger logger.Logger) *PermissionService {
	return &PermissionService{
		storage: storage.PermissionStorage{
			DB:  store.DB,
			Ctx: ctx,
		},
		store:  store,
		ctx:    ctx,
		logger: logger,
	}
}
