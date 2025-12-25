package types

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ExtensionResponse is the typed response for single extension operations
type ExtensionResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    shared_types.Extension `json:"data"`
}

// CategoriesResponse is the typed response for listing categories
type CategoriesResponse struct {
	Status  string                           `json:"status"`
	Message string                           `json:"message"`
	Data    []shared_types.ExtensionCategory `json:"data"`
}

// ExecutionResponse is the typed response for single execution
type ExecutionResponse struct {
	Status  string                           `json:"status"`
	Message string                           `json:"message"`
	Data    *shared_types.ExtensionExecution `json:"data"`
}

// ListExecutionsResponse is the typed response for listing executions
type ListExecutionsResponse struct {
	Status  string                            `json:"status"`
	Message string                            `json:"message"`
	Data    []shared_types.ExtensionExecution `json:"data"`
}

// ListLogsResponseData contains log data with pagination info
type ListLogsResponseData struct {
	Logs            []shared_types.ExtensionLog   `json:"logs"`
	NextAfter       int64                         `json:"next_after"`
	ExecutionStatus *shared_types.ExecutionStatus `json:"execution_status,omitempty"`
}

// ListLogsResponse is the typed response for listing logs
type ListLogsResponse struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Data    ListLogsResponseData `json:"data"`
}

// ListExtensionsResponse wraps the extension list response
type ListExtensionsResponse struct {
	Status  string                             `json:"status"`
	Message string                             `json:"message"`
	Data    shared_types.ExtensionListResponse `json:"data"`
}
