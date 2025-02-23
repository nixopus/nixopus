package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/role/service"
	"github.com/raghavyuva/nixopus-api/internal/features/role/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type RolesController struct {
	store     *shared_storage.Store
	validator *validation.Validator
	service   *service.RoleService
	ctx       context.Context
}

func NewRolesController(
	store *shared_storage.Store,
	ctx context.Context,
) *RolesController {
	return &RolesController{
		store:     store,
		validator: validation.NewValidator(),
		service:   service.NewRoleService(store, ctx),
		ctx:       ctx,
	}
}
