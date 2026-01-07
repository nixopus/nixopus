package tools

import (
	"context"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	mcp_middleware "github.com/raghavyuva/nixopus-api/internal/mcp/middleware"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// RestartDeploymentHandler returns the handler function for restarting a deployment
// Auth middleware is applied automatically during registration
func RestartDeploymentHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	taskService *tasks.TaskService,
) func(context.Context, *mcp.CallToolRequest, RestartDeploymentInput) (*mcp.CallToolResult, RestartDeploymentOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input RestartDeploymentInput,
	) (*mcp.CallToolResult, RestartDeploymentOutput, error) {
		deploymentID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, RestartDeploymentOutput{}, err
		}

		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero RestartDeploymentOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, RestartDeploymentOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero RestartDeploymentOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		userID := user.ID

		restartRequest := types.RestartDeploymentRequest{
			ID: deploymentID,
		}

		err = taskService.RestartDeployment(&restartRequest, userID, organizationID)
		if err != nil {
			l.Log(logger.Error, "Failed to restart deployment", err.Error())
			return nil, RestartDeploymentOutput{}, err
		}

		return nil, RestartDeploymentOutput{
			Response: types.MessageResponse{
				Status:  "success",
				Message: "Application restarted successfully",
			},
		}, nil
	}
}
