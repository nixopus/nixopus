package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/permission/service"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type PermissionController struct {
	service   service.PermissionService
	store     shared_storage.Store
	storage   storage.PermissionStorage
	ctx       context.Context
	validator *validation.Validator
}

func NewPermissionController(store *shared_storage.Store, ctx context.Context) *PermissionController {
	return &PermissionController{
		service:   *service.NewPermissionService(store, ctx),
		store:     *store,
		storage:   storage.PermissionStorage{DB: store.DB, Ctx: ctx},
		ctx:       ctx,
		validator: validation.NewValidator(),
	}
}
