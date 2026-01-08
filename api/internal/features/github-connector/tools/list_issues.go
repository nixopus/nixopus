package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	mcp_middleware "github.com/raghavyuva/nixopus-api/internal/mcp/middleware"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// ListIssuesHandler returns the handler function for listing GitHub issues
// Auth middleware is applied automatically during registration
func ListIssuesHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	githubService *service.GithubConnectorService,
) func(context.Context, *mcp.CallToolRequest, ListIssuesInput) (*mcp.CallToolResult, ListIssuesOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input ListIssuesInput,
	) (*mcp.CallToolResult, ListIssuesOutput, error) {
		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero ListIssuesOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}

		listRequest := types.ListIssuesRequest{
			RepositoryOwner: input.RepositoryOwner,
			RepositoryName:  input.RepositoryName,
			State:           input.State,
			Labels:          input.Labels,
			Assignee:        input.Assignee,
			Creator:         input.Creator,
			Page:            input.Page,
			PerPage:         input.PerPage,
			ConnectorID:     input.ConnectorID,
		}

		issues, err := githubService.ListIssues(user.ID.String(), &listRequest)
		if err != nil {
			l.Log(logger.Error, "Failed to list issues", err.Error())
			var zero ListIssuesOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}

		page := input.Page
		if page == 0 {
			page = 1
		}
		perPage := input.PerPage
		if perPage == 0 {
			perPage = 30
		}

		return nil, ListIssuesOutput{
			Response: types.ListIssuesResponse{
				Status:  "success",
				Message: "Issues fetched successfully",
				Data: types.ListIssuesResponseData{
					Issues:  issues,
					Page:    page,
					PerPage: perPage,
				},
			},
		}, nil
	}
}
