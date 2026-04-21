package types

import (
	shared_types "github.com/nixopus/nixopus/api/internal/types"
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

// ListExtensionsResponse wraps the extension list response
type ListExtensionsResponse struct {
	Status  string                             `json:"status"`
	Message string                             `json:"message"`
	Data    shared_types.ExtensionListResponse `json:"data"`
}
