package tools

import (
	"context"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	mcp_middleware "github.com/raghavyuva/nixopus-api/internal/mcp/middleware"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// GetApplicationHandler returns the handler function for getting an application
// Auth middleware is applied automatically during registration
func GetApplicationHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	deployService *service.DeployService,
) func(context.Context, *mcp.CallToolRequest, GetApplicationInput) (*mcp.CallToolResult, GetApplicationOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetApplicationInput,
	) (*mcp.CallToolResult, GetApplicationOutput, error) {
		applicationID := input.ID

		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero GetApplicationOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, GetApplicationOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero GetApplicationOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		_ = user.ID

		application, err := deployService.GetApplicationById(applicationID, organizationID)
		if err != nil {
			l.Log(logger.Error, "Failed to get application", err.Error())
			var zero GetApplicationOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "application not found or access denied"},
				},
			}, zero, nil
		}

		return nil, GetApplicationOutput{
			Response: types.ApplicationResponse{
				Status:  "success",
				Message: "Application retrieved successfully",
				Data:    application,
			},
		}, nil
	}
}
