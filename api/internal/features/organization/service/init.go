package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"

	cache "github.com/raghavyuva/nixopus-api/internal/cache"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type OrganizationService struct {
	store        *shared_storage.Store
	storage      storage.OrganizationRepository
	user_storage user_storage.UserStorage
	role_storage role_storage.RoleStorage
	Ctx          context.Context
	logger       logger.Logger
	cache        *cache.Cache
}

func NewOrganizationService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, organizationRepository storage.OrganizationRepository, cache *cache.Cache) *OrganizationService {
	return &OrganizationService{
		store:   store,
		storage: organizationRepository,
		logger:  logger,
		user_storage: user_storage.UserStorage{
			DB:  store.DB,
			Ctx: ctx,
		},
		role_storage: role_storage.RoleStorage{DB: store.DB, Ctx: ctx},
		cache:        cache,
		Ctx:          ctx,
	}
}
