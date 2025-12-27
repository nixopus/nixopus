package types

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// OrganizationResponse is the typed response for single organization operations
type OrganizationResponse struct {
	Status  string                    `json:"status"`
	Message string                    `json:"message"`
	Data    shared_types.Organization `json:"data"`
}

// ListOrganizationsResponse is the typed response for listing organizations
type ListOrganizationsResponse struct {
	Status  string                      `json:"status"`
	Message string                      `json:"message"`
	Data    []shared_types.Organization `json:"data"`
}

// OrganizationUsersResponse is the typed response for listing organization users
type OrganizationUsersResponse struct {
	Status  string                                    `json:"status"`
	Message string                                    `json:"message"`
	Data    []shared_types.OrganizationUsersWithRoles `json:"data"`
}

// InviteResponseData holds the data for invite responses
type InviteResponseData struct {
	Email          string `json:"email"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
}

// InviteResponse is the typed response for invite operations
type InviteResponse struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Data    InviteResponseData `json:"data"`
}

// OrganizationSettingsResponse is the typed response for organization settings
type OrganizationSettingsResponse struct {
	Status  string                             `json:"status"`
	Message string                             `json:"message"`
	Data    *shared_types.OrganizationSettings `json:"data"`
}
