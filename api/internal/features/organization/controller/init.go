package controller

import (
	"context"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/cache"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type OrganizationsController struct {
	store        *shared_storage.Store
	validator    *validation.Validator
	service      *service.OrganizationService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
	cache        *cache.Cache
}

func NewOrganizationsController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
	cache *cache.Cache,
) *OrganizationsController {
	storage := storage.OrganizationStore{DB: store.DB, Ctx: ctx}
	return &OrganizationsController{
		store:        store,
		validator:    validation.NewValidator(&storage),
		service:      service.NewOrganizationService(store, ctx, l, &storage, cache),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
		cache:        cache,
	}
}

// Notify sends a notification to the user for the given payload type.
//
// This method constructs a new NotificationPayload object with the given user and request data,
// and sends it to the notification manager.
func (c *OrganizationsController) Notify(payloadType notification.NotificationPayloadType, user *shared_types.User, r *http.Request, data any) {
	c.notification.SendNotification(notification.NewNotificationPayload(
		payloadType,
		user.ID.String(),
		data,
		notification.NotificationCategoryOrganization,
	))
}
