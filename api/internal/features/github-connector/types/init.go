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

// CreateIssueRequest represents the request to create a GitHub issue
type CreateIssueRequest struct {
	RepositoryOwner string   `json:"repository_owner"`
	RepositoryName  string   `json:"repository_name"`
	Title           string   `json:"title"`
	Body            string   `json:"body"`
	Labels          []string `json:"labels,omitempty"`
	Assignees       []string `json:"assignees,omitempty"`
	ConnectorID     string   `json:"connector_id,omitempty"`
}

// UpdateIssueRequest represents the request to update a GitHub issue
type UpdateIssueRequest struct {
	RepositoryOwner string   `json:"repository_owner"`
	RepositoryName  string   `json:"repository_name"`
	IssueNumber     int      `json:"issue_number"`
	Title           *string  `json:"title,omitempty"`
	Body            *string  `json:"body,omitempty"`
	State           *string  `json:"state,omitempty"` // "open" or "closed"
	Labels          []string `json:"labels,omitempty"`
	Assignees       []string `json:"assignees,omitempty"`
	ConnectorID     string   `json:"connector_id,omitempty"`
}

// CommentOnIssueRequest represents the request to comment on a GitHub issue
type CommentOnIssueRequest struct {
	RepositoryOwner string `json:"repository_owner"`
	RepositoryName  string `json:"repository_name"`
	IssueNumber     int    `json:"issue_number"`
	Body            string `json:"body"`
	ConnectorID     string `json:"connector_id,omitempty"`
}

// ListIssuesRequest represents the request to list GitHub issues
type ListIssuesRequest struct {
	RepositoryOwner string   `json:"repository_owner"`
	RepositoryName  string   `json:"repository_name"`
	State           string   `json:"state,omitempty"`    // "open", "closed", or "all"
	Labels          []string `json:"labels,omitempty"`   // Filter by labels
	Assignee        string   `json:"assignee,omitempty"` // Filter by assignee
	Creator         string   `json:"creator,omitempty"`  // Filter by creator
	Page            int      `json:"page,omitempty"`     // Page number (default: 1)
	PerPage         int      `json:"per_page,omitempty"` // Items per page (default: 30, max: 100)
	ConnectorID     string   `json:"connector_id,omitempty"`
}

// IssueResponse represents a GitHub issue response
type IssueResponse struct {
	ID      int    `json:"id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
	User    struct {
		Login string `json:"login"`
	} `json:"user"`
}

// IssueCommentResponse represents a GitHub issue comment response
type IssueCommentResponse struct {
	ID      int    `json:"id"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	User    struct {
		Login string `json:"login"`
	} `json:"user"`
}

// CreateIssueResponse is the typed response for creating an issue
type CreateIssueResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Data    IssueResponse `json:"data"`
}

// UpdateIssueResponse is the typed response for updating an issue
type UpdateIssueResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Data    IssueResponse `json:"data"`
}

// CommentOnIssueResponse is the typed response for commenting on an issue
type CommentOnIssueResponse struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Data    IssueCommentResponse `json:"data"`
}

// ListIssuesResponseData contains the issues data with pagination
type ListIssuesResponseData struct {
	Issues     []IssueResponse `json:"issues"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalCount int             `json:"total_count,omitempty"`
}

// ListIssuesResponse is the typed response for listing issues
type ListIssuesResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    ListIssuesResponseData `json:"data"`
}

var (
	ErrMissingSlug            = errors.New("slug is required")
	ErrMissingPem             = errors.New("pem is required")
	ErrMissingClientID        = errors.New("client_id is required")
	ErrMissingClientSecret    = errors.New("client_secret is required")
	ErrMissingWebhookSecret   = errors.New("webhook_secret is required")
	ErrMissingInstallationID  = errors.New("installation_id is required")
	ErrMissingID              = errors.New("id is required")
	ErrInvalidRequestType     = errors.New("invalid request type")
	ErrConnectorDoesNotExist  = errors.New("connector does not exist")
	ErrNoConnectors           = errors.New("no connectors found")
	ErrPermissionDenied       = errors.New("permission denied")
	ErrMissingRepositoryOwner = errors.New("repository_owner is required")
	ErrMissingRepositoryName  = errors.New("repository_name is required")
	ErrMissingTitle           = errors.New("title is required")
	ErrMissingBody            = errors.New("body is required")
	ErrMissingIssueNumber     = errors.New("issue_number is required")
)
