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

// ListImagesInput is the input structure for the MCP tool
type ListImagesInput struct {
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	All            *bool  `json:"all,omitempty"`
	ContainerID    string `json:"container_id,omitempty"`
	ImagePrefix    string `json:"image_prefix,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i ListImagesInput) GetOrganizationID() string {
	return i.OrganizationID
}

// ListImagesOutput is the output structure for the MCP tool
type ListImagesOutput struct {
	Response container_types.ListImagesResponse `json:"response"`
}

// PruneImagesInput is the input structure for the MCP tool
type PruneImagesInput struct {
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	Until          string `json:"until,omitempty"`
	Label          string `json:"label,omitempty"`
	Dangling       *bool  `json:"dangling,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i PruneImagesInput) GetOrganizationID() string {
	return i.OrganizationID
}

// PruneImagesOutput is the output structure for the MCP tool
type PruneImagesOutput struct {
	Response container_types.PruneImagesResponse `json:"response"`
}

// PruneBuildCacheInput is the input structure for the MCP tool
type PruneBuildCacheInput struct {
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	All            *bool  `json:"all,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i PruneBuildCacheInput) GetOrganizationID() string {
	return i.OrganizationID
}

// PruneBuildCacheOutput is the output structure for the MCP tool
type PruneBuildCacheOutput struct {
	Response container_types.MessageResponse `json:"response"`
}

// RemoveContainerInput is the input structure for the MCP tool
type RemoveContainerInput struct {
	ID             string `json:"id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	Force          *bool  `json:"force,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i RemoveContainerInput) GetOrganizationID() string {
	return i.OrganizationID
}

// RemoveContainerOutput is the output structure for the MCP tool
type RemoveContainerOutput struct {
	Response container_types.ContainerActionResponse `json:"response"`
}

// RestartContainerInput is the input structure for the MCP tool
type RestartContainerInput struct {
	ID             string `json:"id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	Timeout        *int   `json:"timeout,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i RestartContainerInput) GetOrganizationID() string {
	return i.OrganizationID
}

// RestartContainerOutput is the output structure for the MCP tool
type RestartContainerOutput struct {
	Response container_types.ContainerActionResponse `json:"response"`
}

// StartContainerInput is the input structure for the MCP tool
type StartContainerInput struct {
	ID             string `json:"id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i StartContainerInput) GetOrganizationID() string {
	return i.OrganizationID
}

// StartContainerOutput is the output structure for the MCP tool
type StartContainerOutput struct {
	Response container_types.ContainerActionResponse `json:"response"`
}

// StopContainerInput is the input structure for the MCP tool
type StopContainerInput struct {
	ID             string `json:"id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	Timeout        *int   `json:"timeout,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i StopContainerInput) GetOrganizationID() string {
	return i.OrganizationID
}

// StopContainerOutput is the output structure for the MCP tool
type StopContainerOutput struct {
	Response container_types.ContainerActionResponse `json:"response"`
}
