package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type UserService struct {
	storage storage.UserStorage
	Ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewUserService(store *shared_storage.Store, ctx context.Context, logger logger.Logger) *UserService {
	return &UserService{
		storage: storage.UserStorage{
			DB:  store.DB,
			Ctx: ctx,
		},
		store:  store,
		Ctx:    ctx,
		logger: logger,
	}
}
