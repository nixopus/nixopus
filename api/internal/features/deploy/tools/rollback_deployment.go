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

// RollbackDeploymentHandler returns the handler function for rolling back a deployment
// Auth middleware is applied automatically during registration
func RollbackDeploymentHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	taskService *tasks.TaskService,
) func(context.Context, *mcp.CallToolRequest, RollbackDeploymentInput) (*mcp.CallToolResult, RollbackDeploymentOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input RollbackDeploymentInput,
	) (*mcp.CallToolResult, RollbackDeploymentOutput, error) {
		deploymentID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, RollbackDeploymentOutput{}, err
		}

		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero RollbackDeploymentOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, RollbackDeploymentOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero RollbackDeploymentOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		userID := user.ID

		rollbackRequest := types.RollbackDeploymentRequest{
			ID: deploymentID,
		}

		err = taskService.RollbackDeployment(&rollbackRequest, userID, organizationID)
		if err != nil {
			l.Log(logger.Error, "Failed to rollback deployment", err.Error())
			return nil, RollbackDeploymentOutput{}, err
		}

		return nil, RollbackDeploymentOutput{
			Response: types.MessageResponse{
				Status:  "success",
				Message: "Application rolled back successfully",
			},
		}, nil
	}
}
