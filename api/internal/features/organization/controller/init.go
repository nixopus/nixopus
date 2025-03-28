package controller

import (
	"context"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
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
	storage := storage.OrganizationStore{DB: store.DB, Ctx: ctx}
	return &OrganizationsController{
		store:        store,
		validator:    validation.NewValidator(&storage),
		service:      service.NewOrganizationService(store, ctx, l, &storage),
		ctx:          ctx,
		notification: notificationManager,
	}
}

// parseAndValidate parses and validates the request body.
//
// This method attempts to parse the request body into the provided 'req' interface
// using the controller's validator. If parsing fails, an error response is sent
// and the method returns false. It also validates the parsed request object and
// returns false if validation fails. If both operations are successful, it returns true.
//
// Parameters:
//
//	w - the HTTP response writer to send error responses.
//	r - the HTTP request containing the body to parse.
//	req - the interface to populate with the parsed request body.
//
// Returns:
//
//	bool - true if parsing and validation succeed, false otherwise.
func (c *OrganizationsController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	if err := c.validator.ParseRequestBody(r, r.Body, req); err != nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return false
	}

	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

// Notify sends a notification to the user for the given payload type.
//
// This method constructs a new NotificationPayload object with the given user and request data,
// and sends it to the notification manager.
func (c *OrganizationsController) Notify(payloadType notification.NotificationPayloadType, user *shared_types.User, r *http.Request) {
	c.notification.SendNotification(notification.NewNotificationPayload(
		payloadType,
		user.ID.String(),
		notification.NotificationAuthenticationData{
			Email: user.Email,
			NotificationBaseData: notification.NotificationBaseData{
				IP:      r.RemoteAddr,
				Browser: r.UserAgent(),
			},
			UserName: user.Username,
		},
		notification.NotificationCategoryOrganization,
	))
}
