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

// UpdateIssueHandler returns the handler function for updating a GitHub issue
// Auth middleware is applied automatically during registration
func UpdateIssueHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	githubService *service.GithubConnectorService,
) func(context.Context, *mcp.CallToolRequest, UpdateIssueInput) (*mcp.CallToolResult, UpdateIssueOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input UpdateIssueInput,
	) (*mcp.CallToolResult, UpdateIssueOutput, error) {
		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero UpdateIssueOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}

		updateRequest := types.UpdateIssueRequest{
			RepositoryOwner: input.RepositoryOwner,
			RepositoryName:  input.RepositoryName,
			IssueNumber:     input.IssueNumber,
			Title:           input.Title,
			Body:            input.Body,
			State:           input.State,
			Labels:          input.Labels,
			Assignees:       input.Assignees,
			ConnectorID:     input.ConnectorID,
		}

		issue, err := githubService.UpdateIssue(user.ID.String(), &updateRequest)
		if err != nil {
			l.Log(logger.Error, "Failed to update issue", err.Error())
			var zero UpdateIssueOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}

		return nil, UpdateIssueOutput{
			Response: types.UpdateIssueResponse{
				Status:  "success",
				Message: "Issue updated successfully",
				Data:    *issue,
			},
		}, nil
	}
}
