package tools

import deploy_types "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"

// DeleteApplicationInput is the input structure for the MCP tool
type DeleteApplicationInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// DeleteApplicationOutput is the output structure for the MCP tool
type DeleteApplicationOutput struct {
	Response deploy_types.MessageResponse `json:"response"`
}

// GetApplicationDeploymentsInput is the input structure for the MCP tool
type GetApplicationDeploymentsInput struct {
	ID       string `json:"id" jsonschema:"required"`
	Page     string `json:"page,omitempty"`
	PageSize string `json:"page_size,omitempty"`
}

// GetApplicationDeploymentsOutput is the output structure for the MCP tool
type GetApplicationDeploymentsOutput struct {
	Response deploy_types.ListDeploymentsResponse `json:"response"`
}

// GetApplicationInput is the input structure for the MCP tool
type GetApplicationInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// GetApplicationOutput is the output structure for the MCP tool
type GetApplicationOutput struct {
	Response deploy_types.ApplicationResponse `json:"response"`
}

// GetApplicationsInput is the input structure for the MCP tool
type GetApplicationsInput struct {
	Page     string `json:"page,omitempty"`
	PageSize string `json:"page_size,omitempty"`
}

// GetApplicationsOutput is the output structure for the MCP tool
type GetApplicationsOutput struct {
	Response deploy_types.ListApplicationsResponse `json:"response"`
}
