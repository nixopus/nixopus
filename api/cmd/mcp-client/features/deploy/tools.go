package deploy

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	client_types "github.com/raghavyuva/nixopus-api/cmd/mcp-client/types"
	"github.com/raghavyuva/nixopus-api/cmd/mcp-client/utils"
)

// ToolHandler handles deploy feature tool calls
type ToolHandler struct{}

// NewToolHandler creates a new deploy tool handler
func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// GetToolParams returns the tool parameters for a given tool name
func (h *ToolHandler) GetToolParams(toolName string) (*mcp.CallToolParams, error) {
	applicationID := os.Getenv("APPLICATION_ID")
	authToken := os.Getenv("AUTH_TOKEN")

	if applicationID == "" {
		applicationID = "test-application-id"
	}
	if authToken == "" {
		fmt.Println("Warning: AUTH_TOKEN not set. Authentication will fail.")
		fmt.Println("   Set AUTH_TOKEN environment variable with a valid API key.")
	}

	var params *mcp.CallToolParams

	switch toolName {
	case "delete_application":
		params = &mcp.CallToolParams{
			Name: "delete_application",
			Arguments: map[string]any{
				"id": applicationID,
			},
		}
	case "get_application_deployments":
		arguments := map[string]any{
			"id": applicationID,
		}
		// Add optional pagination parameters if set
		if page := os.Getenv("PAGE"); page != "" {
			arguments["page"] = page
		}
		if pageSize := os.Getenv("PAGE_SIZE"); pageSize != "" {
			arguments["page_size"] = pageSize
		}
		params = &mcp.CallToolParams{
			Name:      "get_application_deployments",
			Arguments: arguments,
		}
	case "get_application":
		params = &mcp.CallToolParams{
			Name: "get_application",
			Arguments: map[string]any{
				"id": applicationID,
			},
		}
	case "get_applications":
		arguments := map[string]any{}
		// Add optional pagination parameters if set
		if page := os.Getenv("PAGE"); page != "" {
			arguments["page"] = page
		}
		if pageSize := os.Getenv("PAGE_SIZE"); pageSize != "" {
			arguments["page_size"] = pageSize
		}
		params = &mcp.CallToolParams{
			Name:      "get_applications",
			Arguments: arguments,
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

// TestTool tests a deploy tool
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

// GetAvailableTools returns the list of available deploy tools
func (h *ToolHandler) GetAvailableTools() []string {
	return []string{
		"delete_application",
		"get_application_deployments",
		"get_application",
		"get_applications",
	}
}

// GetToolDescription returns the description for a tool
func (h *ToolHandler) GetToolDescription(toolName string) string {
	descriptions := map[string]string{
		"delete_application":          "Delete a deployed application. Requires application ID.",
		"get_application_deployments": "Get deployments for an application with pagination. Requires application ID. Optionally specify page and page_size.",
		"get_application":             "Get a single application by ID. Requires application ID.",
		"get_applications":            "Get all applications with pagination. Optionally specify page and page_size.",
	}
	return descriptions[toolName]
}
