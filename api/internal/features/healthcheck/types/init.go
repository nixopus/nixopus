package types

import (
	"errors"
)

// CreateHealthCheckRequest represents a request to create a health check
type CreateHealthCheckRequest struct {
	ApplicationID    string            `json:"application_id" validate:"required,uuid"`
	Endpoint         string            `json:"endpoint"`
	Method           string            `json:"method"`
	ExpectedStatus   []int             `json:"expected_status_codes,omitempty"`
	TimeoutSeconds   int               `json:"timeout_seconds,omitempty"`
	IntervalSeconds  int               `json:"interval_seconds,omitempty"`
	FailureThreshold int               `json:"failure_threshold,omitempty"`
	SuccessThreshold int               `json:"success_threshold,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	Body             string            `json:"body,omitempty"`
	RetentionDays    int               `json:"retention_days,omitempty"`
}

// UpdateHealthCheckRequest represents a request to update a health check
type UpdateHealthCheckRequest struct {
	ApplicationID    string            `json:"application_id" validate:"required,uuid"`
	Endpoint         string            `json:"endpoint,omitempty"`
	Method           string            `json:"method,omitempty"`
	ExpectedStatus   []int             `json:"expected_status_codes,omitempty"`
	TimeoutSeconds   int               `json:"timeout_seconds,omitempty"`
	IntervalSeconds  int               `json:"interval_seconds,omitempty"`
	FailureThreshold int               `json:"failure_threshold,omitempty"`
	SuccessThreshold int               `json:"success_threshold,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	Body             string            `json:"body,omitempty"`
	RetentionDays    int               `json:"retention_days,omitempty"`
}

// ToggleHealthCheckRequest represents a request to enable/disable a health check
type ToggleHealthCheckRequest struct {
	ApplicationID string `json:"application_id" validate:"required,uuid"`
	Enabled       bool   `json:"enabled"`
}

// GetHealthCheckResultsRequest represents a request to get health check results
type GetHealthCheckResultsRequest struct {
	ApplicationID string `json:"application_id" validate:"required,uuid"`
	Limit         int    `json:"limit,omitempty"`
	StartTime     string `json:"start_time,omitempty"`
	EndTime       string `json:"end_time,omitempty"`
}

// GetHealthCheckStatsRequest represents a request to get health check statistics
type GetHealthCheckStatsRequest struct {
	ApplicationID string `json:"application_id" validate:"required,uuid"`
	Period        string `json:"period,omitempty"`
}

// Domain-specific errors
var (
	ErrHealthCheckNotFound      = errors.New("health check not found")
	ErrInvalidApplicationID     = errors.New("invalid application ID")
	ErrInvalidEndpoint          = errors.New("invalid endpoint")
	ErrInvalidMethod            = errors.New("invalid HTTP method")
	ErrInvalidTimeout           = errors.New("timeout must be between 5 and 120 seconds")
	ErrInvalidInterval          = errors.New("interval must be between 30 and 3600 seconds")
	ErrInvalidThreshold         = errors.New("threshold must be between 1 and 10")
	ErrInvalidRetentionDays     = errors.New("retention days must be between 1 and 365")
	ErrInvalidRequestType       = errors.New("invalid request type")
	ErrHealthCheckAlreadyExists = errors.New("health check already exists for this application")
	ErrPermissionDenied         = errors.New("permission denied")
	ErrRateLimitExceeded        = errors.New("rate limit exceeded")
)
