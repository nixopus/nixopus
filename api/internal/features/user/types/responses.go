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

// UserPreferencesResponse is the typed response for user preferences
type UserPreferencesResponse struct {
	Status  string                        `json:"status"`
	Message string                        `json:"message"`
	Data    *shared_types.UserPreferences `json:"data"`
}

// IsOnboardedResponseData holds the onboarding status
type IsOnboardedResponseData struct {
	IsOnboarded bool `json:"is_onboarded"`
}

// IsOnboardedResponse is the typed response for onboarding status
type IsOnboardedResponse struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message"`
	Data    IsOnboardedResponseData `json:"data"`
}

// MarkOnboardingCompleteResponse is the response for marking onboarding complete
// Returns only the data field as specified in the API requirements
type MarkOnboardingCompleteResponse struct {
	Data IsOnboardedResponseData `json:"data"`
}
