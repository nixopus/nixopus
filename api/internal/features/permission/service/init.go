package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/permission/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type PermissionService struct {
	storage storage.PermissionStorage
	ctx     context.Context
	store   *shared_storage.Store
}

func NewPermissionService(store *shared_storage.Store, ctx context.Context) *PermissionService {
	return &PermissionService{
		storage: storage.PermissionStorage{
			DB:  store.DB,
			Ctx: ctx,
		},
		store:   store,
		ctx:     ctx,
	}
}