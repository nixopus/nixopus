package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/cache"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/trail/service"
	"github.com/raghavyuva/nixopus-api/internal/features/trail/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/trail/types"
	"github.com/raghavyuva/nixopus-api/internal/features/trail/validation"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// TrailController handles HTTP requests for trail provisioning.
type TrailController struct {
	validator *validation.Validator
	service   *service.TrailService
	ctx       context.Context
	logger    logger.Logger
	cache     *cache.Cache
}

// NewTrailController creates a new TrailController instance.
func NewTrailController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	cache *cache.Cache,
) *TrailController {
	trailStorage := storage.NewTrailStorage(store.DB, ctx)

	return &TrailController{
		validator: validation.NewValidator(),
		service:   service.NewTrailService(store, ctx, l, trailStorage),
		ctx:       ctx,
		logger:    l,
		cache:     cache,
	}
}

// mapErrorToStatus maps domain errors to HTTP status codes.
func mapErrorToStatus(err error) int {
	switch {
	case errors.Is(err, types.ErrImageNotAllowed):
		return http.StatusBadRequest
	case errors.Is(err, types.ErrActiveProvisionExists):
		return http.StatusConflict
	case errors.Is(err, types.ErrSystemAtCapacity):
		return http.StatusServiceUnavailable
	case errors.Is(err, types.ErrProvisionNotFound):
		return http.StatusNotFound
	case errors.Is(err, types.ErrInvalidSessionID):
		return http.StatusBadRequest
	case errors.Is(err, types.ErrOrganizationRequired):
		return http.StatusForbidden
	case errors.Is(err, types.ErrInvalidOrganizationID):
		return http.StatusBadRequest
	case errors.Is(err, types.ErrFailedToEnqueueTask):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
