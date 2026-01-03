package tools

import container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"

// GetContainerLogsInput is the input structure for the MCP tool
type GetContainerLogsInput struct {
	ID             string  `json:"id" jsonschema:"required"`
	OrganizationID string  `json:"organization_id" jsonschema:"required"`
	Follow         bool    `json:"follow,omitempty"`
	Tail           *int    `json:"tail,omitempty"`
	Since          *string `json:"since,omitempty"`
	Until          *string `json:"until,omitempty"`
	Stdout         bool    `json:"stdout,omitempty"`
	Stderr         bool    `json:"stderr,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i GetContainerLogsInput) GetOrganizationID() string {
	return i.OrganizationID
}

// GetContainerLogsOutput is the output structure for the MCP tool
type GetContainerLogsOutput struct {
	Logs string `json:"logs"`
}

// GetContainerInput is the input structure for the MCP tool
type GetContainerInput struct {
	ID             string `json:"id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i GetContainerInput) GetOrganizationID() string {
	return i.OrganizationID
}

// GetContainerOutput is the output structure for the MCP tool
type GetContainerOutput struct {
	Container container_types.Container `json:"container"`
}

// ListContainersInput is the input structure for the MCP tool
type ListContainersInput struct {
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	Page           *int   `json:"page,omitempty"`
	PageSize       *int   `json:"page_size,omitempty"`
	Search         string `json:"search,omitempty"`
	SortBy         string `json:"sort_by,omitempty"`
	SortOrder      string `json:"sort_order,omitempty"`
	Status         string `json:"status,omitempty"`
	Name           string `json:"name,omitempty"`
	Image          string `json:"image,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i ListContainersInput) GetOrganizationID() string {
	return i.OrganizationID
}

// ListContainersOutput is the output structure for the MCP tool
type ListContainersOutput struct {
	Response container_types.ListContainersResponse `json:"response"`
}
