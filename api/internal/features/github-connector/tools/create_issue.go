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

// CreateIssueHandler returns the handler function for creating a GitHub issue
// Auth middleware is applied automatically during registration
func CreateIssueHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	githubService *service.GithubConnectorService,
) func(context.Context, *mcp.CallToolRequest, CreateIssueInput) (*mcp.CallToolResult, CreateIssueOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input CreateIssueInput,
	) (*mcp.CallToolResult, CreateIssueOutput, error) {
		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero CreateIssueOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}

		createRequest := types.CreateIssueRequest{
			RepositoryOwner: input.RepositoryOwner,
			RepositoryName:  input.RepositoryName,
			Title:           input.Title,
			Body:            input.Body,
			Labels:          input.Labels,
			Assignees:       input.Assignees,
			ConnectorID:     input.ConnectorID,
		}

		issue, err := githubService.CreateIssue(user.ID.String(), &createRequest)
		if err != nil {
			l.Log(logger.Error, "Failed to create issue", err.Error())
			var zero CreateIssueOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}

		return nil, CreateIssueOutput{
			Response: types.CreateIssueResponse{
				Status:  "success",
				Message: "Issue created successfully",
				Data:    *issue,
			},
		}, nil
	}
}
