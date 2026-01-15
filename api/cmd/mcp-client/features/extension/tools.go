package extension

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	client_types "github.com/raghavyuva/nixopus-api/cmd/mcp-client/types"
	"github.com/raghavyuva/nixopus-api/cmd/mcp-client/utils"
)

// ToolHandler handles extension feature tool calls
type ToolHandler struct{}

// NewToolHandler creates a new extension tool handler
func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// GetToolParams returns the tool parameters for a given tool name
func (h *ToolHandler) GetToolParams(toolName string) (*mcp.CallToolParams, error) {
	extensionID := os.Getenv("EXTENSION_ID")
	executionID := os.Getenv("EXECUTION_ID")
	authToken := os.Getenv("AUTH_TOKEN")
	variablesJSON := os.Getenv("VARIABLES")

	if extensionID == "" {
		extensionID = "test-extension-id"
	}
	if executionID == "" {
		executionID = "test-execution-id"
	}
	if authToken == "" {
		fmt.Println("Warning: AUTH_TOKEN not set. Authentication will fail.")
		fmt.Println("   Set AUTH_TOKEN environment variable with a valid API key.")
	}

	var params *mcp.CallToolParams

	switch toolName {
	case "list_extensions":
		args := map[string]any{
			"page":      1,
			"page_size": 10,
		}
		if category := os.Getenv("CATEGORY"); category != "" {
			args["category"] = category
		}
		if extType := os.Getenv("EXTENSION_TYPE"); extType != "" {
			args["type"] = extType
		}
		if search := os.Getenv("SEARCH"); search != "" {
			args["search"] = search
		}
		params = &mcp.CallToolParams{
			Name:      "list_extensions",
			Arguments: args,
		}
	case "get_extension":
		params = &mcp.CallToolParams{
			Name: "get_extension",
			Arguments: map[string]any{
				"id": extensionID,
			},
		}
	case "run_extension":
		args := map[string]any{
			"extension_id": extensionID,
		}
		if variablesJSON != "" {
			var variables map[string]interface{}
			if err := json.Unmarshal([]byte(variablesJSON), &variables); err == nil {
				args["variables"] = variables
			}
		}
		params = &mcp.CallToolParams{
			Name:      "run_extension",
			Arguments: args,
		}
	case "get_execution":
		params = &mcp.CallToolParams{
			Name: "get_execution",
			Arguments: map[string]any{
				"execution_id": executionID,
			},
		}
	case "list_execution_logs":
		args := map[string]any{
			"execution_id": executionID,
			"limit":        200,
		}
		if afterSeq := os.Getenv("AFTER_SEQ"); afterSeq != "" {
			var seq int64
			if _, err := fmt.Sscanf(afterSeq, "%d", &seq); err == nil {
				args["after_seq"] = seq
			}
		}
		params = &mcp.CallToolParams{
			Name:      "list_execution_logs",
			Arguments: args,
		}
	case "cancel_execution":
		params = &mcp.CallToolParams{
			Name: "cancel_execution",
			Arguments: map[string]any{
				"execution_id": executionID,
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

// TestTool tests an extension tool
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

// GetAvailableTools returns the list of available extension tools
func (h *ToolHandler) GetAvailableTools() []string {
	return []string{
		"list_extensions",
		"get_extension",
		"run_extension",
		"get_execution",
		"list_execution_logs",
		"cancel_execution",
	}
}

// GetToolDescription returns the description for a tool
func (h *ToolHandler) GetToolDescription(toolName string) string {
	descriptions := map[string]string{
		"list_extensions":     "List extensions with pagination, filtering, and sorting",
		"get_extension":       "Get detailed information about an extension",
		"run_extension":       "Run an extension with variable values",
		"get_execution":       "Get execution status and details",
		"list_execution_logs": "List execution logs with pagination",
		"cancel_execution":    "Cancel a running extension execution",
	}
	return descriptions[toolName]
}
