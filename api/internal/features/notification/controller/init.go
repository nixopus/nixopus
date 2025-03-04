package controller

import (
	"context"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/service"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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
func (c *NotificationController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
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

// GetUser retrieves the current user from the request context.
//
// This method extracts the user from the request context using the UserContextKey.
// If the user cannot be retrieved or is of the wrong type, an error is logged,
// an error response is sent to the client, and the method returns nil.
//
// Parameters:
//
//	w - the HTTP response writer used to send error responses.
//	r - the HTTP request containing the context from which to retrieve the user.
//
// Returns:
//
//	*shared_types.User - a pointer to the User object retrieved from the context,
//	or nil if the user could not be retrieved.
func (c *NotificationController) GetUser(w http.ResponseWriter, r *http.Request) *shared_types.User {
	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		c.logger.Log(logger.Error, shared_types.ErrFailedToGetUserFromContext.Error(), shared_types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return nil
	}

	return user
}
