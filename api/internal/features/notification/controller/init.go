package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/service"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type NotificationController struct {
	store        *shared_storage.Store
	validator    *validation.Validator
	service      *service.NotificationService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

// NewNotificationController creates a new NotificationController with the given App.
//
// This function creates a new NotificationController with the given App and returns a pointer to it.
//
// The App passed to this function should be a valid App that has been created with storage.NewApp.
func NewNotificationController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *NotificationController {
	return &NotificationController{
		store:        store,
		validator:    validation.NewValidator(),
		service:      service.NewNotificationService(store, ctx, l),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
	}
}
