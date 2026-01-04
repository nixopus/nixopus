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

// RedeployApplicationHandler returns the handler function for redeploying an application
// Auth middleware is applied automatically during registration
func RedeployApplicationHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	taskService *tasks.TaskService,
) func(context.Context, *mcp.CallToolRequest, RedeployApplicationInput) (*mcp.CallToolResult, RedeployApplicationOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input RedeployApplicationInput,
	) (*mcp.CallToolResult, RedeployApplicationOutput, error) {
		applicationID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, RedeployApplicationOutput{}, err
		}

		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero RedeployApplicationOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, RedeployApplicationOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero RedeployApplicationOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		userID := user.ID

		redeployRequest := types.ReDeployApplicationRequest{
			ID:                applicationID,
			Force:             input.Force,
			ForceWithoutCache: input.ForceWithoutCache,
		}

		application, err := taskService.ReDeployApplication(&redeployRequest, userID, organizationID)
		if err != nil {
			l.Log(logger.Error, "Failed to redeploy application", err.Error())
			return nil, RedeployApplicationOutput{}, err
		}

		return nil, RedeployApplicationOutput{
			Response: types.ApplicationResponse{
				Status:  "success",
				Message: "Application redeployed successfully",
				Data:    application,
			},
		}, nil
	}
}
