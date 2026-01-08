package tools

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	mcp_middleware "github.com/raghavyuva/nixopus-api/internal/mcp/middleware"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// GetApplicationsHandler returns the handler function for getting applications
// Auth middleware is applied automatically during registration
func GetApplicationsHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	deployService *service.DeployService,
) func(context.Context, *mcp.CallToolRequest, GetApplicationsInput) (*mcp.CallToolResult, GetApplicationsOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetApplicationsInput,
	) (*mcp.CallToolResult, GetApplicationsOutput, error) {
		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero GetApplicationsOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, GetApplicationsOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero GetApplicationsOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		_ = user.ID

		page := "1"
		if input.Page != "" {
			if p, err := strconv.Atoi(input.Page); err == nil && p > 0 {
				page = input.Page
			}
		}

		pageSize := "10"
		if input.PageSize != "" {
			if ps, err := strconv.Atoi(input.PageSize); err == nil && ps > 0 {
				pageSize = input.PageSize
			}
		}

		applications, totalCount, err := deployService.GetApplications(page, pageSize, organizationID)
		if err != nil {
			l.Log(logger.Error, "Failed to get applications", err.Error())
			return nil, GetApplicationsOutput{}, err
		}

		return nil, GetApplicationsOutput{
			Response: types.ListApplicationsResponse{
				Status:  "success",
				Message: "Applications retrieved successfully",
				Data: types.ListApplicationsResponseData{
					Applications: applications,
					TotalCount:   totalCount,
					Page:         page,
					PageSize:     pageSize,
				},
			},
		}, nil
	}
}
