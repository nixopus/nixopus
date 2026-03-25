package service

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/storage"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
)

type UserService struct {
	storage storage.UserRepository
	Ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewUserService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, storage storage.UserRepository) *UserService {
	return &UserService{
		storage: storage,
		store:   store,
		Ctx:     ctx,
		logger:  logger,
	}
}
