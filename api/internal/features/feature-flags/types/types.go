package types

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ListFeatureFlagsResponse is the typed response for listing feature flags
type ListFeatureFlagsResponse struct {
	Status  string                     `json:"status"`
	Message string                     `json:"message"`
	Data    []shared_types.FeatureFlag `json:"data"`
}

// IsFeatureEnabledData contains the is_enabled flag
type IsFeatureEnabledData struct {
	IsEnabled bool `json:"is_enabled"`
}

// IsFeatureEnabledResponse is the typed response for checking if a feature is enabled
type IsFeatureEnabledResponse struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Data    IsFeatureEnabledData `json:"data"`
}
