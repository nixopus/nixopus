package controller

import (
	"context"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/extension/service"
	"github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/extension/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type ExtensionsController struct {
	store     *shared_storage.Store
	service   *service.ExtensionService
	validator *validation.Validator
	ctx       context.Context
	logger    logger.Logger
}

func NewExtensionsController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) *ExtensionsController {
	storage := storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
	return &ExtensionsController{
		store:     store,
		service:   service.NewExtensionService(store, ctx, l, &storage),
		validator: validation.NewValidator(&storage),
		ctx:       ctx,
		logger:    l,
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
func (c *ExtensionsController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	if err := c.validator.ParseRequestBody(r, r.Body, req); err != nil {
		c.logger.Log(logger.Error, "Failed to decode request", err.Error())
		utils.SendErrorResponse(w, "Failed to decode request", http.StatusBadRequest)
		return false
	}

	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

type RunExtensionRequest struct {
	Variables map[string]interface{} `json:"variables"`
}
