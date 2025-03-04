package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type PermissionService struct {
	storage storage.PermissionRepository
	ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewPermissionService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, permissionRepository storage.PermissionRepository) *PermissionService {
	return &PermissionService{
		storage: permissionRepository,
		store:   store,
		ctx:     ctx,
		logger:  logger,
	}
}
