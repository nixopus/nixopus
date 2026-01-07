package container

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	client_types "github.com/raghavyuva/nixopus-api/cmd/mcp-client/types"
	"github.com/raghavyuva/nixopus-api/cmd/mcp-client/utils"
)

// ToolHandler handles container feature tool calls
type ToolHandler struct{}

// NewToolHandler creates a new container tool handler
func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// GetToolParams returns the tool parameters for a given tool name
func (h *ToolHandler) GetToolParams(toolName string) (*mcp.CallToolParams, error) {
	containerID := os.Getenv("CONTAINER_ID")
	authToken := os.Getenv("AUTH_TOKEN")

	if containerID == "" {
		containerID = "test-container-id"
	}
	if authToken == "" {
		fmt.Println("Warning: AUTH_TOKEN not set. Authentication will fail.")
		fmt.Println("   Set AUTH_TOKEN environment variable with a valid API key.")
	}

	var params *mcp.CallToolParams

	switch toolName {
	case "get_container":
		params = &mcp.CallToolParams{
			Name: "get_container",
			Arguments: map[string]any{
				"id": containerID,
			},
		}
	case "get_container_logs":
		params = &mcp.CallToolParams{
			Name: "get_container_logs",
			Arguments: map[string]any{
				"id":     containerID,
				"follow": false,
				"tail":   100,
				"stdout": true,
				"stderr": true,
			},
		}
	case "list_containers":
		params = &mcp.CallToolParams{
			Name: "list_containers",
			Arguments: map[string]any{
				"page":       1,
				"page_size":  10,
				"sort_by":    "name",
				"sort_order": "asc",
			},
		}
	case "list_images":
		params = &mcp.CallToolParams{
			Name: "list_images",
			Arguments: map[string]any{
				"all": false,
			},
		}
	case "prune_images":
		params = &mcp.CallToolParams{
			Name: "prune_images",
			Arguments: map[string]any{
				"dangling": true,
			},
		}
	case "prune_build_cache":
		params = &mcp.CallToolParams{
			Name: "prune_build_cache",
			Arguments: map[string]any{
				"all": true,
			},
		}
	case "remove_container":
		params = &mcp.CallToolParams{
			Name: "remove_container",
			Arguments: map[string]any{
				"id":    containerID,
				"force": true,
			},
		}
	case "restart_container":
		params = &mcp.CallToolParams{
			Name: "restart_container",
			Arguments: map[string]any{
				"id":      containerID,
				"timeout": 10,
			},
		}
	case "start_container":
		params = &mcp.CallToolParams{
			Name: "start_container",
			Arguments: map[string]any{
				"id": containerID,
			},
		}
	case "stop_container":
		params = &mcp.CallToolParams{
			Name: "stop_container",
			Arguments: map[string]any{
				"id":      containerID,
				"timeout": 10,
			},
		}
	case "update_container_resources":
		params = &mcp.CallToolParams{
			Name: "update_container_resources",
			Arguments: map[string]any{
				"id":          containerID,
				"memory":      1073741824, // 1GB in bytes
				"memory_swap": 2147483648, // 2GB in bytes
				"cpu_shares":  1024,
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

// TestTool tests a container tool
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

// GetAvailableTools returns the list of available container tools
func (h *ToolHandler) GetAvailableTools() []string {
	return []string{
		"get_container",
		"get_container_logs",
		"list_containers",
		"list_images",
		"prune_images",
		"prune_build_cache",
		"remove_container",
		"restart_container",
		"start_container",
		"stop_container",
		"update_container_resources",
	}
}

// GetToolDescription returns the description for a tool
func (h *ToolHandler) GetToolDescription(toolName string) string {
	descriptions := map[string]string{
		"get_container":              "Get detailed information about a Docker container",
		"get_container_logs":         "Get logs from a Docker container",
		"list_containers":            "List Docker containers with pagination, filtering, and sorting",
		"list_images":                "List Docker images with optional filtering",
		"prune_images":               "Prune Docker images with optional filtering by until time, label, or dangling status",
		"prune_build_cache":          "Prune Docker build cache, optionally prune all cache entries",
		"remove_container":           "Remove a Docker container, optionally force removal",
		"restart_container":          "Restart a Docker container, optionally specify timeout in seconds",
		"start_container":            "Start a Docker container",
		"stop_container":             "Stop a Docker container, optionally specify timeout in seconds",
		"update_container_resources": "Update resource limits (memory, memory swap, CPU shares) of a running Docker container",
	}
	return descriptions[toolName]
}
