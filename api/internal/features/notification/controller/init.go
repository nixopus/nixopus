package controller

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/service"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type NotificationController struct {
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
	storage := storage.NotificationStorage{DB: store.DB, Ctx: ctx}
	return &NotificationController{
		validator:    validation.NewValidator(&storage),
		service:      service.NewNotificationService(store, ctx, l, &storage),
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
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to read request body", err.Error())
		return false
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.validator.ParseRequestBody(r, io.NopCloser(bytes.NewBuffer(bodyBytes)), req); err != nil {
		c.logger.Log(logger.Error, "Failed to decode request", err.Error())
		return false
	}

	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		return false
	}

	return true
}
