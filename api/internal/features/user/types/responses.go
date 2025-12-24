package types

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// UserResponse is the typed response for user details
type UserResponse struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Data    *shared_types.User `json:"data"`
}

// UserOrganizationsListResponse is the typed response for listing user organizations
type UserOrganizationsListResponse struct {
	Status  string                      `json:"status"`
	Message string                      `json:"message"`
	Data    []UserOrganizationsResponse `json:"data"`
}

// UserSettingsResponse is the typed response for user settings
type UserSettingsResponse struct {
	Status  string                     `json:"status"`
	Message string                     `json:"message"`
	Data    *shared_types.UserSettings `json:"data"`
}

// UpdateUsernameResponseData holds the updated username
type UpdateUsernameResponseData struct {
	Name string `json:"name"`
}

// UpdateUsernameResponse is the typed response for username update
type UpdateUsernameResponse struct {
	Status  string                     `json:"status"`
	Message string                     `json:"message"`
	Data    UpdateUsernameResponseData `json:"data"`
}
