package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"

	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type OrganizationService struct {
	store   *shared_storage.Store
	storage storage.OrganizationStore
	Ctx     context.Context
}

func NewOrganizationService(store *shared_storage.Store, ctx context.Context) *OrganizationService {
	return &OrganizationService{
		store: store,
		storage: storage.OrganizationStore{
			DB:  store.DB,
			Ctx: ctx,
		},
	}
}