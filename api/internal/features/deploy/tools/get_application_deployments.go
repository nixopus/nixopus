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

// GetApplicationDeploymentsHandler returns the handler function for getting application deployments
// Auth middleware is applied automatically during registration
func GetApplicationDeploymentsHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	deployService *service.DeployService,
) func(context.Context, *mcp.CallToolRequest, GetApplicationDeploymentsInput) (*mcp.CallToolResult, GetApplicationDeploymentsOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetApplicationDeploymentsInput,
	) (*mcp.CallToolResult, GetApplicationDeploymentsOutput, error) {
		applicationID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, GetApplicationDeploymentsOutput{}, err
		}

		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero GetApplicationDeploymentsOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, GetApplicationDeploymentsOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero GetApplicationDeploymentsOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		_ = user.ID

		_, err = deployService.GetApplicationById(applicationID.String(), organizationID)
		if err != nil {
			var zero GetApplicationDeploymentsOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "application not found or access denied"},
				},
			}, zero, nil
		}

		page := 1
		if input.Page != "" {
			if p, err := strconv.Atoi(input.Page); err == nil && p > 0 {
				page = p
			}
		}

		pageSize := 10
		if input.PageSize != "" {
			if ps, err := strconv.Atoi(input.PageSize); err == nil && ps > 0 {
				pageSize = ps
			}
		}

		deployments, totalCount, err := deployService.GetApplicationDeployments(applicationID, page, pageSize)
		if err != nil {
			l.Log(logger.Error, "Failed to get application deployments", err.Error())
			return nil, GetApplicationDeploymentsOutput{}, err
		}

		return nil, GetApplicationDeploymentsOutput{
			Response: types.ListDeploymentsResponse{
				Status:  "success",
				Message: "Application deployments retrieved successfully",
				Data: types.ListDeploymentsResponseData{
					Deployments: deployments,
					TotalCount:  totalCount,
					Page:        strconv.Itoa(page),
					PageSize:    strconv.Itoa(pageSize),
				},
			},
		}, nil
	}
}
