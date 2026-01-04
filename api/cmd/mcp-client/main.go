package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/cmd/mcp-client/features"
)

func main() {
	ctx := context.Background()

	// Get the path to the MCP server binary
	serverPath := os.Getenv("MCP_SERVER_PATH")
	if serverPath == "" {
		log.Fatalf("MCP_SERVER_PATH environment variable is required.\nExample: export MCP_SERVER_PATH=./nixopus-mcp-server")
	}

	// Create a new MCP client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "nixopus-mcp-client",
		Version: "1.0.0",
	}, nil)

	// Connect to the server over stdin/stdout using CommandTransport
	transport := &mcp.CommandTransport{
		Command: exec.Command(serverPath),
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("Failed to connect to MCP server: %v", err)
	}
	defer session.Close()

	fmt.Println("Connected to MCP server")

	// List available tools
	fmt.Println("\nListing available tools...")
	toolsResult, err := session.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	fmt.Printf("Found %d tool(s):\n", len(toolsResult.Tools))
	for _, tool := range toolsResult.Tools {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}

	// Initialize feature registry
	registry := features.InitializeRegistry()

	// Test tools if test argument is provided
	if len(os.Args) > 1 && os.Args[1] == "test" {
		toolName := os.Getenv("TOOL_NAME")
		if toolName == "" {
			toolName = "get_container_logs" // Default tool
		}

		if err := registry.TestTool(ctx, session, toolName); err != nil {
			log.Fatalf("Test failed: %v", err)
		}
	} else {
		printUsage(registry)
	}
}

func printUsage(registry *features.Registry) {
	fmt.Println("\nUsage:")
	fmt.Println("  Set MCP_SERVER_PATH environment variable:")
	fmt.Println("    export MCP_SERVER_PATH=./nixopus-mcp-server")
	fmt.Println("")
	fmt.Println("  To test a tool call, run:")
	fmt.Println("    CONTAINER_ID=<id> AUTH_TOKEN=<token> go run ./cmd/mcp-client test")
	fmt.Println("")
	fmt.Println("  To test a specific tool, set TOOL_NAME:")
	fmt.Println("    TOOL_NAME=get_container CONTAINER_ID=<id> AUTH_TOKEN=<token> go run ./cmd/mcp-client test")
	fmt.Println("    TOOL_NAME=get_container_logs CONTAINER_ID=<id> AUTH_TOKEN=<token> go run ./cmd/mcp-client test")
	fmt.Println("")
	fmt.Println("  Available tools:")

	// Get all available tools from registry
	containerHandler, _ := registry.GetHandler("container")
	if containerHandler != nil {
		for _, toolName := range containerHandler.GetAvailableTools() {
			description := containerHandler.GetToolDescription(toolName)
			fmt.Printf("    - %s: %s\n", toolName, description)
		}
	}

	fmt.Println("")
	fmt.Println("  Note: AUTH_TOKEN is required for authentication")
}
