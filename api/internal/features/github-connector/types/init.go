package types

import (
	"errors"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

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
