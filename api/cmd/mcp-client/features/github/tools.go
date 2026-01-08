package github

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	client_types "github.com/raghavyuva/nixopus-api/cmd/mcp-client/types"
	"github.com/raghavyuva/nixopus-api/cmd/mcp-client/utils"
)

// ToolHandler handles GitHub issue feature tool calls
type ToolHandler struct{}

// NewToolHandler creates a new GitHub tool handler
func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// GetToolParams returns the tool parameters for a given tool name
func (h *ToolHandler) GetToolParams(toolName string) (*mcp.CallToolParams, error) {
	authToken := os.Getenv("AUTH_TOKEN")
	repoOwner := os.Getenv("REPO_OWNER")
	repoName := os.Getenv("REPO_NAME")
	connectorID := os.Getenv("CONNECTOR_ID")

	if authToken == "" {
		fmt.Println("Warning: AUTH_TOKEN not set. Authentication will fail.")
		fmt.Println("   Set AUTH_TOKEN environment variable with a valid API key.")
	}

	var params *mcp.CallToolParams

	switch toolName {
	case "create_issue":
		title := os.Getenv("ISSUE_TITLE")
		body := os.Getenv("ISSUE_BODY")
		labels := os.Getenv("ISSUE_LABELS")
		assignees := os.Getenv("ISSUE_ASSIGNEES")

		if repoOwner == "" || repoName == "" || title == "" || body == "" {
			return nil, fmt.Errorf("REPO_OWNER, REPO_NAME, ISSUE_TITLE, and ISSUE_BODY are required")
		}

		args := map[string]any{
			"repository_owner": repoOwner,
			"repository_name":  repoName,
			"title":            title,
			"body":             body,
		}

		if labels != "" {
			args["labels"] = strings.Split(labels, ",")
		}
		if assignees != "" {
			args["assignees"] = strings.Split(assignees, ",")
		}
		if connectorID != "" {
			args["connector_id"] = connectorID
		}

		params = &mcp.CallToolParams{
			Name:      "create_issue",
			Arguments: args,
		}

	case "update_issue":
		issueNumberStr := os.Getenv("ISSUE_NUMBER")
		title := os.Getenv("ISSUE_TITLE")
		body := os.Getenv("ISSUE_BODY")
		state := os.Getenv("ISSUE_STATE")
		labels := os.Getenv("ISSUE_LABELS")
		assignees := os.Getenv("ISSUE_ASSIGNEES")

		if repoOwner == "" || repoName == "" || issueNumberStr == "" {
			return nil, fmt.Errorf("REPO_OWNER, REPO_NAME, and ISSUE_NUMBER are required")
		}

		issueNumber, err := strconv.Atoi(issueNumberStr)
		if err != nil {
			return nil, fmt.Errorf("ISSUE_NUMBER must be a valid integer: %w", err)
		}

		args := map[string]any{
			"repository_owner": repoOwner,
			"repository_name":  repoName,
			"issue_number":     issueNumber,
		}

		if title != "" {
			args["title"] = title
		}
		if body != "" {
			args["body"] = body
		}
		if state != "" {
			args["state"] = state
		}
		if labels != "" {
			args["labels"] = strings.Split(labels, ",")
		}
		if assignees != "" {
			args["assignees"] = strings.Split(assignees, ",")
		}
		if connectorID != "" {
			args["connector_id"] = connectorID
		}

		params = &mcp.CallToolParams{
			Name:      "update_issue",
			Arguments: args,
		}

	case "comment_on_issue":
		issueNumberStr := os.Getenv("ISSUE_NUMBER")
		body := os.Getenv("COMMENT_BODY")

		if repoOwner == "" || repoName == "" || issueNumberStr == "" || body == "" {
			return nil, fmt.Errorf("REPO_OWNER, REPO_NAME, ISSUE_NUMBER, and COMMENT_BODY are required")
		}

		issueNumber, err := strconv.Atoi(issueNumberStr)
		if err != nil {
			return nil, fmt.Errorf("ISSUE_NUMBER must be a valid integer: %w", err)
		}

		args := map[string]any{
			"repository_owner": repoOwner,
			"repository_name":  repoName,
			"issue_number":     issueNumber,
			"body":             body,
		}

		if connectorID != "" {
			args["connector_id"] = connectorID
		}

		params = &mcp.CallToolParams{
			Name:      "comment_on_issue",
			Arguments: args,
		}

	case "list_issues":
		state := os.Getenv("ISSUE_STATE")
		labels := os.Getenv("ISSUE_LABELS")
		assignee := os.Getenv("ISSUE_ASSIGNEE")
		creator := os.Getenv("ISSUE_CREATOR")
		pageStr := os.Getenv("PAGE")
		perPageStr := os.Getenv("PER_PAGE")

		if repoOwner == "" || repoName == "" {
			return nil, fmt.Errorf("REPO_OWNER and REPO_NAME are required")
		}

		args := map[string]any{
			"repository_owner": repoOwner,
			"repository_name":  repoName,
		}

		if state != "" {
			args["state"] = state
		}
		if labels != "" {
			args["labels"] = strings.Split(labels, ",")
		}
		if assignee != "" {
			args["assignee"] = assignee
		}
		if creator != "" {
			args["creator"] = creator
		}
		if pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err == nil {
				args["page"] = page
			}
		}
		if perPageStr != "" {
			if perPage, err := strconv.Atoi(perPageStr); err == nil {
				args["per_page"] = perPage
			}
		}
		if connectorID != "" {
			args["connector_id"] = connectorID
		}

		params = &mcp.CallToolParams{
			Name:      "list_issues",
			Arguments: args,
		}

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	// Add auth token to metadata if provided
	if authToken != "" {
		params.Meta = mcp.Meta{
			"auth_token": authToken,
		}
	}

	return params, nil
}

// TestTool tests a GitHub tool
func (h *ToolHandler) TestTool(ctx context.Context, session client_types.Session, toolName string) error {
	fmt.Printf("\nTesting %s tool...\n", toolName)

	params, err := h.GetToolParams(toolName)
	if err != nil {
		return err
	}

	res, err := session.CallTool(ctx, params)
	if err != nil {
		return fmt.Errorf("CallTool failed: %w", err)
	}

	utils.PrintToolResponse(res)

	if res.IsError {
		return fmt.Errorf("tool returned an error")
	}

	return nil
}

// GetAvailableTools returns the list of available GitHub tools
func (h *ToolHandler) GetAvailableTools() []string {
	return []string{
		"create_issue",
		"update_issue",
		"comment_on_issue",
		"list_issues",
	}
}

// GetToolDescription returns the description for a tool
func (h *ToolHandler) GetToolDescription(toolName string) string {
	descriptions := map[string]string{
		"create_issue":     "Create a new GitHub issue in a repository",
		"update_issue":     "Update an existing GitHub issue",
		"comment_on_issue": "Add a comment to a GitHub issue",
		"list_issues":      "List GitHub issues for a repository",
	}

	if desc, ok := descriptions[toolName]; ok {
		return desc
	}
	return "Unknown tool"
}
