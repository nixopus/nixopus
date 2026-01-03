package tools

import shared_types "github.com/raghavyuva/nixopus-api/internal/types"

// ListExtensionsInput is the input structure for the MCP tool
type ListExtensionsInput struct {
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	Category       string `json:"category,omitempty"`
	Type           string `json:"type,omitempty"`
	Search         string `json:"search,omitempty"`
	SortBy         string `json:"sort_by,omitempty"`
	SortDir        string `json:"sort_dir,omitempty"`
	Page           *int   `json:"page,omitempty"`
	PageSize       *int   `json:"page_size,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i ListExtensionsInput) GetOrganizationID() string {
	return i.OrganizationID
}

// ListExtensionsOutput is the output structure for the MCP tool
type ListExtensionsOutput struct {
	Response shared_types.ExtensionListResponse `json:"response"`
}

// GetExtensionInput is the input structure for the MCP tool
type GetExtensionInput struct {
	ID             string `json:"id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i GetExtensionInput) GetOrganizationID() string {
	return i.OrganizationID
}

// GetExtensionOutput is the output structure for the MCP tool
type GetExtensionOutput struct {
	Extension shared_types.Extension `json:"extension"`
}

// RunExtensionInput is the input structure for the MCP tool
type RunExtensionInput struct {
	ExtensionID    string                 `json:"extension_id" jsonschema:"required"`
	OrganizationID string                 `json:"organization_id" jsonschema:"required"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i RunExtensionInput) GetOrganizationID() string {
	return i.OrganizationID
}

// RunExtensionOutput is the output structure for the MCP tool
type RunExtensionOutput struct {
	Execution shared_types.ExtensionExecution `json:"execution"`
}

// GetExecutionInput is the input structure for the MCP tool
type GetExecutionInput struct {
	ExecutionID    string `json:"execution_id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i GetExecutionInput) GetOrganizationID() string {
	return i.OrganizationID
}

// GetExecutionOutput is the output structure for the MCP tool
type GetExecutionOutput struct {
	Execution shared_types.ExtensionExecution `json:"execution"`
}

// ListExecutionLogsInput is the input structure for the MCP tool
type ListExecutionLogsInput struct {
	ExecutionID    string `json:"execution_id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	AfterSeq       *int64 `json:"after_seq,omitempty"`
	Limit          *int   `json:"limit,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i ListExecutionLogsInput) GetOrganizationID() string {
	return i.OrganizationID
}

// ListExecutionLogsOutput is the output structure for the MCP tool
type ListExecutionLogsOutput struct {
	Logs            []shared_types.ExtensionLog   `json:"logs"`
	NextAfter       int64                         `json:"next_after"`
	ExecutionStatus *shared_types.ExecutionStatus `json:"execution_status,omitempty"`
}

// CancelExecutionInput is the input structure for the MCP tool
type CancelExecutionInput struct {
	ExecutionID    string `json:"execution_id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i CancelExecutionInput) GetOrganizationID() string {
	return i.OrganizationID
}

// CancelExecutionOutput is the output structure for the MCP tool
type CancelExecutionOutput struct {
	Message string `json:"message"`
}
