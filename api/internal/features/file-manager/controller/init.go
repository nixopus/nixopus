package controller

import (
	"context"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type FileManagerController struct {
	validator    *validation.Validator
	service      *service.FileManagerService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewFileManagerController(
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *FileManagerController {
	return &FileManagerController{
		validator:    validation.NewValidator(),
		service:      service.NewFileManagerService(ctx, l),
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
func (c *FileManagerController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	user := utils.GetUser(w, r)

	if user == nil {
		return false
	}

	if err := c.validator.ParseRequestBody(r, r.Body, req); err != nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return false
	}

	if err := c.validator.ValidateRequest(req, *user); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}
