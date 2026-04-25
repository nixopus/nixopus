package types

import (
	"errors"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// CreateHealthCheckRequest represents a request to create a health check
type CreateHealthCheckRequest struct {
	ApplicationID    string            `json:"application_id" validate:"required,uuid" description:"ID of the application to monitor" example:"550e8400-e29b-41d4-a716-446655440000"`
	Endpoint         string            `json:"endpoint" description:"Health check endpoint path or URL. Defaults to /" example:"/health"`
	Method           string            `json:"method" description:"HTTP method to use. Must be GET, POST, or HEAD. Defaults to GET" example:"GET"`
	ExpectedStatus   []int             `json:"expected_status_codes,omitempty" description:"Expected HTTP status codes. Defaults to [200]"`
	TimeoutSeconds   int               `json:"timeout_seconds,omitempty" validate:"omitempty,min=5,max=120" description:"Request timeout in seconds (5-120). Defaults to 30" example:"30"`
	IntervalSeconds  int               `json:"interval_seconds,omitempty" validate:"omitempty,min=30,max=3600" description:"Check interval in seconds (30-3600). Defaults to 60" example:"60"`
	FailureThreshold int               `json:"failure_threshold,omitempty" validate:"omitempty,min=1,max=10" description:"Consecutive failures before marking unhealthy (1-10). Defaults to 3" example:"3"`
	SuccessThreshold int               `json:"success_threshold,omitempty" validate:"omitempty,min=1,max=10" description:"Consecutive successes before marking healthy (1-10). Defaults to 1" example:"1"`
	Headers          map[string]string `json:"headers,omitempty" description:"Custom HTTP headers to include in health check requests"`
	Body             string            `json:"body,omitempty" description:"Request body to send with health check requests"`
	RetentionDays    int               `json:"retention_days,omitempty" validate:"omitempty,min=1,max=365" description:"Number of days to retain health check results (1-365). Defaults to 30" example:"30"`
}

// UpdateHealthCheckRequest represents a request to update a health check
type UpdateHealthCheckRequest struct {
	ApplicationID    string            `json:"application_id" validate:"required,uuid" description:"ID of the application to update health check for" example:"550e8400-e29b-41d4-a716-446655440000"`
	Endpoint         string            `json:"endpoint,omitempty" description:"Health check endpoint path or URL. Must start with /, http://, or https://" example:"/health"`
	Method           string            `json:"method,omitempty" description:"HTTP method to use. Must be GET, POST, or HEAD" example:"GET"`
	ExpectedStatus   []int             `json:"expected_status_codes,omitempty" description:"Expected HTTP status codes"`
	TimeoutSeconds   int               `json:"timeout_seconds,omitempty" validate:"omitempty,min=5,max=120" description:"Request timeout in seconds (5-120)" example:"30"`
	IntervalSeconds  int               `json:"interval_seconds,omitempty" validate:"omitempty,min=30,max=3600" description:"Check interval in seconds (30-3600)" example:"60"`
	FailureThreshold int               `json:"failure_threshold,omitempty" validate:"omitempty,min=1,max=10" description:"Consecutive failures before marking unhealthy (1-10)" example:"3"`
	SuccessThreshold int               `json:"success_threshold,omitempty" validate:"omitempty,min=1,max=10" description:"Consecutive successes before marking healthy (1-10)" example:"1"`
	Headers          map[string]string `json:"headers,omitempty" description:"Custom HTTP headers to include in health check requests"`
	Body             string            `json:"body,omitempty" description:"Request body to send with health check requests"`
	RetentionDays    int               `json:"retention_days,omitempty" validate:"omitempty,min=1,max=365" description:"Number of days to retain health check results (1-365)" example:"30"`
}

// ToggleHealthCheckRequest represents a request to enable/disable a health check
type ToggleHealthCheckRequest struct {
	ApplicationID string `json:"application_id" validate:"required,uuid" description:"ID of the application to toggle health check for" example:"550e8400-e29b-41d4-a716-446655440000"`
	Enabled       bool   `json:"enabled" description:"Whether the health check should be enabled"`
}

// GetHealthCheckResultsRequest represents a request to get health check results
type GetHealthCheckResultsRequest struct {
	ApplicationID string `json:"application_id" validate:"required,uuid" description:"ID of the application to get results for" example:"550e8400-e29b-41d4-a716-446655440000"`
	Limit         int    `json:"limit,omitempty" validate:"omitempty,min=0,max=1000" description:"Maximum number of results to return (0-1000). Defaults to 100" example:"100"`
	StartTime     string `json:"start_time,omitempty" description:"Filter results after this time (RFC 3339 format)" example:"2024-01-01T00:00:00Z"`
	EndTime       string `json:"end_time,omitempty" description:"Filter results before this time (RFC 3339 format)" example:"2024-12-31T23:59:59Z"`
}

// GetHealthCheckStatsRequest represents a request to get health check statistics
type GetHealthCheckStatsRequest struct {
	ApplicationID string `json:"application_id" validate:"required,uuid" description:"ID of the application to get statistics for" example:"550e8400-e29b-41d4-a716-446655440000"`
	Period        string `json:"period,omitempty" description:"Time period for statistics aggregation" example:"24h"`
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

// HealthCheckResponse is a typed response for single health check operations.
type HealthCheckResponse struct {
	Status  string                    `json:"status"`
	Message string                    `json:"message,omitempty"`
	Data    *shared_types.HealthCheck `json:"data,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// HealthCheckResultsResponse is a typed response for health check results.
type HealthCheckResultsResponse struct {
	Status  string                            `json:"status"`
	Message string                            `json:"message,omitempty"`
	Data    []*shared_types.HealthCheckResult `json:"data,omitempty"`
	Error   string                            `json:"error,omitempty"`
}

// HealthCheckStatsData is the typed stats payload for health check metrics.
type HealthCheckStatsData struct {
	ApplicationID    string  `json:"application_id"`
	UptimePercentage float64 `json:"uptime_percentage"`
	AvgResponseTime  int     `json:"avg_response_time_ms"`
	TotalChecks      int     `json:"total_checks"`
	SuccessfulChecks int     `json:"successful_checks"`
	FailedChecks     int     `json:"failed_checks"`
	Period           string  `json:"period"`
	LastStatus       string  `json:"last_status"`
}

// HealthCheckStatsResponse is a typed response for health check statistics.
type HealthCheckStatsResponse struct {
	Status  string                `json:"status"`
	Message string                `json:"message,omitempty"`
	Data    *HealthCheckStatsData `json:"data,omitempty"`
	Error   string                `json:"error,omitempty"`
}

// HealthCheckMessageResponse is a typed message-only response.
type HealthCheckMessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
