package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/role/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type RoleService struct {
	store   *shared_storage.Store
	storage storage.RoleStorage
	Ctx     context.Context
	logger  logger.Logger
}

func NewRoleService(store *shared_storage.Store, ctx context.Context, logger logger.Logger) *RoleService {
	return &RoleService{
		store:   store,
		storage: storage.RoleStorage{DB: store.DB, Ctx: ctx},
		Ctx:     ctx,
		logger:  logger,
	}
}
