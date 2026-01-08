package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// GetDeploymentByIdHandler returns the handler function for getting a deployment by ID
// Auth middleware is applied automatically during registration
func GetDeploymentByIdHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	deployService *service.DeployService,
) func(context.Context, *mcp.CallToolRequest, GetDeploymentByIdInput) (*mcp.CallToolResult, GetDeploymentByIdOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetDeploymentByIdInput,
	) (*mcp.CallToolResult, GetDeploymentByIdOutput, error) {
		deploymentID := input.ID

		deployment, err := deployService.GetDeploymentById(deploymentID)
		if err != nil {
			l.Log(logger.Error, "Failed to get deployment", err.Error())
			var zero GetDeploymentByIdOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "deployment not found or access denied"},
				},
			}, zero, nil
		}

		return nil, GetDeploymentByIdOutput{
			Response: types.DeploymentResponse{
				Status:  "success",
				Message: "Deployment retrieved successfully",
				Data:    deployment,
			},
		}, nil
	}
}
