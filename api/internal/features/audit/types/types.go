package types

import shared_types "github.com/raghavyuva/nixopus-api/internal/types"

// ActivityMessage represents a human-readable activity from audit logs
// @Description Human-readable activity derived from audit log
type ActivityMessage struct {
	// Unique identifier for the activity
	ID string `json:"id"`
	// Human-readable message describing the activity
	Message string `json:"message"`
	// Action type (create, update, delete, access)
	Action shared_types.AuditAction `json:"action"`
	// Actor who performed the action (username or email)
	Actor string `json:"actor"`
	// Resource type that was acted upon
	Resource string `json:"resource"`
	// ID of the resource that was acted upon
	ResourceID string `json:"resource_id"`
	// ISO 8601 timestamp when the action occurred
	Timestamp string `json:"timestamp"`
	// Additional metadata about the action
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Color associated with the action type for UI purposes
	ActionColor string `json:"action_color"`
}

// PaginationInfo contains pagination details for list responses
// @Description Pagination information for list responses
type PaginationInfo struct {
	// Current page number (1-indexed)
	CurrentPage int `json:"current_page"`
	// Number of items per page
	PageSize int `json:"page_size"`
	// Total number of items across all pages
	TotalCount int `json:"total_count"`
	// Total number of pages
	TotalPages int `json:"total_pages"`
	// Whether there is a next page
	HasNext bool `json:"has_next"`
	// Whether there is a previous page
	HasPrev bool `json:"has_prev"`
}

// GetActivitiesResponseData contains the data returned when fetching activities
// @Description Response data for GetRecentAuditLogs endpoint
type GetActivitiesResponseData struct {
	// List of activities
	Activities []*ActivityMessage `json:"activities"`
	// Pagination information
	Pagination PaginationInfo `json:"pagination"`
}

// GetActivitiesResponse is the full typed response for the GetRecentAuditLogs endpoint
// @Description Response for GetRecentAuditLogs endpoint
type GetActivitiesResponse struct {
	// Status of the response ("success" or "error")
	Status string `json:"status"`
	// Message providing additional information
	Message string `json:"message,omitempty"`
	// Response data containing activities and pagination
	Data GetActivitiesResponseData `json:"data"`
}

// ErrorResponse represents an error response
// @Description Error response structure
type ErrorResponse struct {
	// Status will always be "error"
	Status string `json:"status"`
	// Error message describing what went wrong
	Error string `json:"error"`
}
