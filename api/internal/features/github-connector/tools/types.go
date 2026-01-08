package tools

import github_types "github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"

// CreateIssueInput is the input structure for the MCP tool
type CreateIssueInput struct {
	RepositoryOwner string   `json:"repository_owner" jsonschema:"required"`
	RepositoryName  string   `json:"repository_name" jsonschema:"required"`
	Title           string   `json:"title" jsonschema:"required"`
	Body            string   `json:"body" jsonschema:"required"`
	Labels          []string `json:"labels,omitempty"`
	Assignees       []string `json:"assignees,omitempty"`
	ConnectorID     string   `json:"connector_id,omitempty"`
}

// CreateIssueOutput is the output structure for the MCP tool
type CreateIssueOutput struct {
	Response github_types.CreateIssueResponse `json:"response"`
}

// UpdateIssueInput is the input structure for the MCP tool
type UpdateIssueInput struct {
	RepositoryOwner string   `json:"repository_owner" jsonschema:"required"`
	RepositoryName  string   `json:"repository_name" jsonschema:"required"`
	IssueNumber     int      `json:"issue_number" jsonschema:"required"`
	Title           *string  `json:"title,omitempty"`
	Body            *string  `json:"body,omitempty"`
	State           *string  `json:"state,omitempty"`
	Labels          []string `json:"labels,omitempty"`
	Assignees       []string `json:"assignees,omitempty"`
	ConnectorID     string   `json:"connector_id,omitempty"`
}

// UpdateIssueOutput is the output structure for the MCP tool
type UpdateIssueOutput struct {
	Response github_types.UpdateIssueResponse `json:"response"`
}

// CommentOnIssueInput is the input structure for the MCP tool
type CommentOnIssueInput struct {
	RepositoryOwner string `json:"repository_owner" jsonschema:"required"`
	RepositoryName  string `json:"repository_name" jsonschema:"required"`
	IssueNumber     int    `json:"issue_number" jsonschema:"required"`
	Body            string `json:"body" jsonschema:"required"`
	ConnectorID     string `json:"connector_id,omitempty"`
}

// CommentOnIssueOutput is the output structure for the MCP tool
type CommentOnIssueOutput struct {
	Response github_types.CommentOnIssueResponse `json:"response"`
}

// ListIssuesInput is the input structure for the MCP tool
type ListIssuesInput struct {
	RepositoryOwner string   `json:"repository_owner" jsonschema:"required"`
	RepositoryName  string   `json:"repository_name" jsonschema:"required"`
	State           string   `json:"state,omitempty"`
	Labels          []string `json:"labels,omitempty"`
	Assignee        string   `json:"assignee,omitempty"`
	Creator         string   `json:"creator,omitempty"`
	Page            int      `json:"page,omitempty"`
	PerPage         int      `json:"per_page,omitempty"`
	ConnectorID     string   `json:"connector_id,omitempty"`
}

// ListIssuesOutput is the output structure for the MCP tool
type ListIssuesOutput struct {
	Response github_types.ListIssuesResponse `json:"response"`
}
