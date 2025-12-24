package types

import (
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// SMTPConfigResponse is the typed response for SMTP config operations
type SMTPConfigResponse struct {
	Status  string                    `json:"status"`
	Message string                    `json:"message"`
	Data    *shared_types.SMTPConfigs `json:"data"`
}

// WebhookConfigResponse is the typed response for webhook config operations
type WebhookConfigResponse struct {
	Status  string                      `json:"status"`
	Message string                      `json:"message"`
	Data    *shared_types.WebhookConfig `json:"data"`
}

// PreferencesResponse is the typed response for preferences
type PreferencesResponse struct {
	Status  string                               `json:"status"`
	Message string                               `json:"message"`
	Data    *notification.GetPreferencesResponse `json:"data"`
}
