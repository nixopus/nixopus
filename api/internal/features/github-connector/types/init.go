package types

import (
	"errors"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type CreateGithubConnectorRequest struct {
	AppID          string `json:"app_id" description:"GitHub App ID. Required if providing custom app credentials" example:"123456"`
	Slug           string `json:"slug" description:"GitHub App slug" example:"my-github-app"`
	Pem            string `json:"pem" description:"GitHub App private key in PEM format"`
	ClientID       string `json:"client_id" description:"GitHub App OAuth client ID" example:"Iv1.abc123def456"`
	ClientSecret   string `json:"client_secret" description:"GitHub App OAuth client secret"`
	WebhookSecret  string `json:"webhook_secret" description:"GitHub App webhook secret"`
	InstallationID string `json:"installation_id,omitempty" description:"GitHub App installation ID" example:"12345678"`
}

type UpdateGithubConnectorRequest struct {
	InstallationID string `json:"installation_id" validate:"required" description:"GitHub App installation ID" example:"12345678"`
	ConnectorID    string `json:"connector_id,omitempty" description:"Connector ID to update. If provided, updates this specific connector" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ListConnectorsResponse is the typed response for listing connectors
type ListConnectorsResponse struct {
	Status  string                         `json:"status"`
	Message string                         `json:"message"`
	Data    []shared_types.GithubConnector `json:"data"`
}

// ListRepositoriesResponseData contains the repositories data with pagination
type ListRepositoriesResponseData struct {
	Repositories []shared_types.GithubRepository `json:"repositories"`
	TotalCount   int                             `json:"total_count"`
	Page         int                             `json:"page"`
	PageSize     int                             `json:"page_size"`
}

// ListRepositoriesResponse is the typed response for listing repositories
type ListRepositoriesResponse struct {
	Status  string                       `json:"status"`
	Message string                       `json:"message"`
	Data    ListRepositoriesResponseData `json:"data"`
}

// ListBranchesResponse is the typed response for listing branches
type ListBranchesResponse struct {
	Status  string                                `json:"status"`
	Message string                                `json:"message"`
	Data    []shared_types.GithubRepositoryBranch `json:"data"`
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
