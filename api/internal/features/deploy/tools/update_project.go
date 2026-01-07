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
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// UpdateProjectHandler returns the handler function for updating a project
// Auth middleware is applied automatically during registration
func UpdateProjectHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	taskService *tasks.TaskService,
) func(context.Context, *mcp.CallToolRequest, UpdateProjectInput) (*mcp.CallToolResult, UpdateProjectOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input UpdateProjectInput,
	) (*mcp.CallToolResult, UpdateProjectOutput, error) {
		applicationID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, UpdateProjectOutput{}, err
		}

		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero UpdateProjectOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, UpdateProjectOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero UpdateProjectOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		userID := user.ID

		// Convert string environment to proper type if provided
		var environment shared_types.Environment
		if input.Environment != "" {
			environment = shared_types.Environment(input.Environment)
		}

		updateRequest := types.UpdateDeploymentRequest{
			ID:                   applicationID,
			Name:                 input.Name,
			Environment:          environment,
			PreRunCommand:        input.PreRunCommand,
			PostRunCommand:       input.PostRunCommand,
			BuildVariables:       input.BuildVariables,
			EnvironmentVariables: input.EnvironmentVariables,
			Port:                 input.Port,
			Force:                input.Force,
			DockerfilePath:       input.DockerfilePath,
			BasePath:             input.BasePath,
		}

		application, err := taskService.UpdateDeployment(&updateRequest, userID, organizationID)
		if err != nil {
			l.Log(logger.Error, "Failed to update project", err.Error())
			return nil, UpdateProjectOutput{}, err
		}

		return nil, UpdateProjectOutput{
			Response: types.ApplicationResponse{
				Status:  "success",
				Message: "Application updated successfully",
				Data:    application,
			},
		}, nil
	}
}
