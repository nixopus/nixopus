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

// CommentOnIssueHandler returns the handler function for commenting on a GitHub issue
// Auth middleware is applied automatically during registration
func CommentOnIssueHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	githubService *service.GithubConnectorService,
) func(context.Context, *mcp.CallToolRequest, CommentOnIssueInput) (*mcp.CallToolResult, CommentOnIssueOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input CommentOnIssueInput,
	) (*mcp.CallToolResult, CommentOnIssueOutput, error) {
		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero CommentOnIssueOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}

		commentRequest := types.CommentOnIssueRequest{
			RepositoryOwner: input.RepositoryOwner,
			RepositoryName:  input.RepositoryName,
			IssueNumber:     input.IssueNumber,
			Body:            input.Body,
			ConnectorID:     input.ConnectorID,
		}

		comment, err := githubService.CommentOnIssue(user.ID.String(), &commentRequest)
		if err != nil {
			l.Log(logger.Error, "Failed to comment on issue", err.Error())
			var zero CommentOnIssueOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}

		return nil, CommentOnIssueOutput{
			Response: types.CommentOnIssueResponse{
				Status:  "success",
				Message: "Comment added successfully",
				Data:    *comment,
			},
		}, nil
	}
}
