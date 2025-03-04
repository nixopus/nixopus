package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/role/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type RoleService struct {
	store   *shared_storage.Store
	storage storage.RoleRepository
	Ctx     context.Context
	logger  logger.Logger
}

func NewRoleService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, roleRepository storage.RoleRepository) *RoleService {
	return &RoleService{
		store:   store,
		storage: roleRepository,
		Ctx:     ctx,
		logger:  logger,
	}
}
