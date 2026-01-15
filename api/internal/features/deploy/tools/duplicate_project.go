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
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// DuplicateProjectHandler returns the handler function for duplicating a project
// Auth middleware is applied automatically during registration
func DuplicateProjectHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	deployService *service.DeployService,
) func(context.Context, *mcp.CallToolRequest, DuplicateProjectInput) (*mcp.CallToolResult, DuplicateProjectOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input DuplicateProjectInput,
	) (*mcp.CallToolResult, DuplicateProjectOutput, error) {
		sourceProjectID, err := uuid.Parse(input.SourceProjectID)
		if err != nil {
			return nil, DuplicateProjectOutput{}, err
		}

		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero DuplicateProjectOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, DuplicateProjectOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero DuplicateProjectOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		userID := user.ID

		// Convert string environment to proper type
		environment := shared_types.Environment(input.Environment)

		duplicateRequest := types.DuplicateProjectRequest{
			SourceProjectID: sourceProjectID,
			Domain:          input.Domain,
			Environment:     environment,
			Branch:          input.Branch,
		}

		application, err := deployService.DuplicateProject(&duplicateRequest, userID, organizationID)
		if err != nil {
			l.Log(logger.Error, "Failed to duplicate project", err.Error())
			return nil, DuplicateProjectOutput{}, err
		}

		return nil, DuplicateProjectOutput{
			Response: types.ApplicationResponse{
				Status:  "success",
				Message: "Project duplicated successfully",
				Data:    application,
			},
		}, nil
	}
}
