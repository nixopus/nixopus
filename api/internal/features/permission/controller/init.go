package controller

import (
	"context"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/service"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type PermissionController struct {
	service   service.PermissionService
	store     shared_storage.Store
	storage   storage.PermissionStorage
	ctx       context.Context
	validator *validation.Validator
	logger    logger.Logger
}

func NewPermissionController(store *shared_storage.Store, ctx context.Context, logger logger.Logger) *PermissionController {
	storage := storage.PermissionStorage{DB: store.DB, Ctx: ctx}
	return &PermissionController{
		service:   *service.NewPermissionService(store, ctx, logger, &storage),
		store:     *store,
		storage:   storage,
		ctx:       ctx,
		validator: validation.NewValidator(&storage),
		logger:    logger,
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
func (c *PermissionController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
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
