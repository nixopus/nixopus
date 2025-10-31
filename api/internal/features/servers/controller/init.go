package controller

import (
	"context"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/service"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type ServersController struct {
	store        *shared_storage.Store
	validator    *validation.Validator
	service      *service.ServersService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewServersController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *ServersController {
	storage := storage.ServerStorage{DB: store.DB, Ctx: ctx}
	return &ServersController{
		store:        store,
		validator:    validation.NewValidator(&storage),
		service:      service.NewServersService(store, ctx, l, &storage),
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
func (c *ServersController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}
