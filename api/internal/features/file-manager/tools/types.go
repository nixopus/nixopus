package tools

import file_manager_types "github.com/raghavyuva/nixopus-api/internal/features/file-manager/types"

// ListFilesInput is the input structure for the MCP tool
type ListFilesInput struct {
	Path           string `json:"path" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i ListFilesInput) GetOrganizationID() string {
	return i.OrganizationID
}

// ListFilesOutput is the output structure for the MCP tool
type ListFilesOutput struct {
	Response file_manager_types.ListFilesResponse `json:"response"`
}

// CreateDirectoryInput is the input structure for the MCP tool
type CreateDirectoryInput struct {
	Path           string `json:"path" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i CreateDirectoryInput) GetOrganizationID() string {
	return i.OrganizationID
}

// CreateDirectoryOutput is the output structure for the MCP tool
type CreateDirectoryOutput struct {
	Response file_manager_types.MessageResponse `json:"response"`
}

// DeleteFileInput is the input structure for the MCP tool
type DeleteFileInput struct {
	Path           string `json:"path" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i DeleteFileInput) GetOrganizationID() string {
	return i.OrganizationID
}

// DeleteFileOutput is the output structure for the MCP tool
type DeleteFileOutput struct {
	Response file_manager_types.MessageResponse `json:"response"`
}

// MoveFileInput is the input structure for the MCP tool
type MoveFileInput struct {
	FromPath       string `json:"from_path" jsonschema:"required"`
	ToPath         string `json:"to_path" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i MoveFileInput) GetOrganizationID() string {
	return i.OrganizationID
}

// MoveFileOutput is the output structure for the MCP tool
type MoveFileOutput struct {
	Response file_manager_types.MessageResponse `json:"response"`
}

// CopyDirectoryInput is the input structure for the MCP tool
type CopyDirectoryInput struct {
	FromPath       string `json:"from_path" jsonschema:"required"`
	ToPath         string `json:"to_path" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i CopyDirectoryInput) GetOrganizationID() string {
	return i.OrganizationID
}

// CopyDirectoryOutput is the output structure for the MCP tool
type CopyDirectoryOutput struct {
	Response file_manager_types.MessageResponse `json:"response"`
}
