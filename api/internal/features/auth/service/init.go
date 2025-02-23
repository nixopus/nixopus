package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type  AuthService struct {
	storage storage.UserStorage
	Ctx     context.Context
	store   *shared_storage.Store
}

func NewAuthService(store *shared_storage.Store,ctx context.Context) *AuthService {
	return &AuthService{
		storage: storage.UserStorage{
			DB:  store.DB,
			Ctx: ctx,
		},
		store:   store,
		Ctx:     ctx,
	}
}