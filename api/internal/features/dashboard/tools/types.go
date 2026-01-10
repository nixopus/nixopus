package tools

import dashboard_types "github.com/raghavyuva/nixopus-api/internal/features/dashboard"

// GetSystemStatsInput is the input structure for the MCP tool
type GetSystemStatsInput struct {
}

// GetSystemStatsOutput is the output structure for the MCP tool
type GetSystemStatsOutput struct {
	Stats dashboard_types.SystemStats `json:"stats"`
}
