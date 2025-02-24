package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type AuthService struct {
	storage storage.UserStorage
	Ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewAuthService(store *shared_storage.Store, ctx context.Context, logger logger.Logger) *AuthService {
	return &AuthService{
		storage: storage.UserStorage{
			DB:  store.DB,
			Ctx: ctx,
		},
		store:  store,
		Ctx:    ctx,
		logger: logger,
	}
}
