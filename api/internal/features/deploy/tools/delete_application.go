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

// DeleteApplicationHandler returns the handler function for deleting an application
// Auth middleware is applied automatically during registration
func DeleteApplicationHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	taskService *tasks.TaskService,
) func(context.Context, *mcp.CallToolRequest, DeleteApplicationInput) (*mcp.CallToolResult, DeleteApplicationOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input DeleteApplicationInput,
	) (*mcp.CallToolResult, DeleteApplicationOutput, error) {
		applicationID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, DeleteApplicationOutput{}, err
		}

		deleteRequest := types.DeleteDeploymentRequest{
			ID: applicationID,
		}

		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero DeleteApplicationOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, DeleteApplicationOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero DeleteApplicationOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		userID := user.ID

		err = taskService.DeleteDeployment(&deleteRequest, userID, organizationID)
		if err != nil {
			return nil, DeleteApplicationOutput{}, err
		}

		return nil, DeleteApplicationOutput{
			Response: types.MessageResponse{
				Status:  "success",
				Message: "Application deleted successfully",
			},
		}, nil
	}
}
