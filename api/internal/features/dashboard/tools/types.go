package tools

import dashboard_types "github.com/raghavyuva/nixopus-api/internal/features/dashboard"

// GetSystemStatsInput is the input structure for the MCP tool
type GetSystemStatsInput struct {
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i GetSystemStatsInput) GetOrganizationID() string {
	return i.OrganizationID
}

// GetSystemStatsOutput is the output structure for the MCP tool
type GetSystemStatsOutput struct {
	Stats dashboard_types.SystemStats `json:"stats"`
}
