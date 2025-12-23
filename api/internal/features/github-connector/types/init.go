package types

import "errors"

type CreateGithubConnectorRequest struct {
	AppID         string `json:"app_id"`
	Slug          string `json:"slug"`
	Pem           string `json:"pem"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	WebhookSecret string `json:"webhook_secret"`
}

type UpdateGithubConnectorRequest struct {
	InstallationID string `json:"installation_id"`
	ConnectorID    string `json:"connector_id,omitempty"` // Optional: if provided, update this specific connector
}

var (
	ErrMissingSlug           = errors.New("slug is required")
	ErrMissingPem            = errors.New("pem is required")
	ErrMissingClientID       = errors.New("client_id is required")
	ErrMissingClientSecret   = errors.New("client_secret is required")
	ErrMissingWebhookSecret  = errors.New("webhook_secret is required")
	ErrMissingInstallationID = errors.New("installation_id is required")
	ErrMissingID             = errors.New("id is required")
	ErrInvalidRequestType    = errors.New("invalid request type")
	ErrConnectorDoesNotExist = errors.New("connector does not exist")
	ErrNoConnectors          = errors.New("no connectors found")
	ErrPermissionDenied      = errors.New("permission denied")
)
