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

// DeployProjectHandler returns the handler function for deploying a project
// Auth middleware is applied automatically during registration
func DeployProjectHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	taskService *tasks.TaskService,
) func(context.Context, *mcp.CallToolRequest, DeployProjectInput) (*mcp.CallToolResult, DeployProjectOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input DeployProjectInput,
	) (*mcp.CallToolResult, DeployProjectOutput, error) {
		applicationID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, DeployProjectOutput{}, err
		}

		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero DeployProjectOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, DeployProjectOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero DeployProjectOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		userID := user.ID

		deployRequest := types.DeployProjectRequest{
			ID: applicationID,
		}

		application, err := taskService.DeployProject(&deployRequest, userID, organizationID)
		if err != nil {
			l.Log(logger.Error, "Failed to deploy project", err.Error())
			return nil, DeployProjectOutput{}, err
		}

		return nil, DeployProjectOutput{
			Response: types.ApplicationResponse{
				Status:  "success",
				Message: "Deployment started successfully",
				Data:    application,
			},
		}, nil
	}
}
