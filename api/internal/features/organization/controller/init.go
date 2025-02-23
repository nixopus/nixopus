package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type OrganizationsController struct {
	store     *shared_storage.Store
	validator *validation.Validator
	service   *service.OrganizationService
	ctx       context.Context
}

func NewOrganizationsController(
	store *shared_storage.Store,
	ctx context.Context,
) *OrganizationsController {
	return &OrganizationsController{
		store:     store,
		validator: validation.NewValidator(),
		service:   service.NewOrganizationService(store, ctx),
		ctx:       ctx,
	}
}
