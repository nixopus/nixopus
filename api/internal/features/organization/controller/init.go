package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type OrganizationsController struct {
	store        *shared_storage.Store
	validator    *validation.Validator
	service      *service.OrganizationService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewOrganizationsController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *OrganizationsController {
	return &OrganizationsController{
		store:        store,
		validator:    validation.NewValidator(),
		service:      service.NewOrganizationService(store, ctx, l),
		ctx:          ctx,
		notification: notificationManager,
	}
}
