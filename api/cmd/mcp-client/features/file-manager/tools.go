package file_manager

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	client_types "github.com/raghavyuva/nixopus-api/cmd/mcp-client/types"
	"github.com/raghavyuva/nixopus-api/cmd/mcp-client/utils"
)

// ToolHandler handles file-manager feature tool calls
type ToolHandler struct{}

// NewToolHandler creates a new file-manager tool handler
func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// GetToolParams returns the tool parameters for a given tool name
func (h *ToolHandler) GetToolParams(toolName string) (*mcp.CallToolParams, error) {
	path := os.Getenv("PATH")
	fromPath := os.Getenv("FROM_PATH")
	toPath := os.Getenv("TO_PATH")
	authToken := os.Getenv("AUTH_TOKEN")

	if path == "" {
		path = "/"
	}
	if fromPath == "" {
		fromPath = "/test/source"
	}
	if toPath == "" {
		toPath = "/test/destination"
	}
	if authToken == "" {
		fmt.Println("Warning: AUTH_TOKEN not set. Authentication will fail.")
		fmt.Println("   Set AUTH_TOKEN environment variable with a valid API key.")
	}

	var params *mcp.CallToolParams

	switch toolName {
	case "list_files":
		params = &mcp.CallToolParams{
			Name: "list_files",
			Arguments: map[string]any{
				"path": path,
			},
		}
	case "create_directory":
		params = &mcp.CallToolParams{
			Name: "create_directory",
			Arguments: map[string]any{
				"path": path,
			},
		}
	case "delete_file":
		params = &mcp.CallToolParams{
			Name: "delete_file",
			Arguments: map[string]any{
				"path": path,
			},
		}
	case "move_file":
		params = &mcp.CallToolParams{
			Name: "move_file",
			Arguments: map[string]any{
				"from_path": fromPath,
				"to_path":   toPath,
			},
		}
	case "copy_directory":
		params = &mcp.CallToolParams{
			Name: "copy_directory",
			Arguments: map[string]any{
				"from_path": fromPath,
				"to_path":   toPath,
			},
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

// TestTool tests a file-manager tool
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

// GetAvailableTools returns the list of available file-manager tools
func (h *ToolHandler) GetAvailableTools() []string {
	return []string{
		"list_files",
		"create_directory",
		"delete_file",
		"move_file",
		"copy_directory",
	}
}

// GetToolDescription returns the description for a tool
func (h *ToolHandler) GetToolDescription(toolName string) string {
	descriptions := map[string]string{
		"list_files":       "List files and directories in a given path",
		"create_directory": "Create a new directory at the given path",
		"delete_file":      "Delete a file or directory at the given path",
		"move_file":        "Move or rename a file or directory from one path to another",
		"copy_directory":   "Copy a file or directory from one path to another",
	}
	return descriptions[toolName]
}
