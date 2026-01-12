package controller

import (
	"errors"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
)

// mapHealthCheckError maps domain-specific errors to appropriate HTTP status codes
func mapHealthCheckError(err error) (int, error) {
	if err == nil {
		return http.StatusInternalServerError, err
	}

	switch {
	case errors.Is(err, types.ErrInvalidApplicationID),
		errors.Is(err, types.ErrInvalidEndpoint),
		errors.Is(err, types.ErrInvalidMethod),
		errors.Is(err, types.ErrInvalidTimeout),
		errors.Is(err, types.ErrInvalidInterval),
		errors.Is(err, types.ErrInvalidThreshold),
		errors.Is(err, types.ErrInvalidRetentionDays),
		errors.Is(err, types.ErrInvalidRequestType):
		return http.StatusBadRequest, err
	case errors.Is(err, types.ErrHealthCheckNotFound):
		return http.StatusNotFound, err
	case errors.Is(err, types.ErrHealthCheckAlreadyExists):
		return http.StatusConflict, err
	case errors.Is(err, types.ErrPermissionDenied):
		return http.StatusForbidden, err
	case errors.Is(err, types.ErrRateLimitExceeded):
		return http.StatusTooManyRequests, err
	default:
		return http.StatusInternalServerError, err
	}
}
