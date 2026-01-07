package types

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Session interface for MCP session operations
type Session interface {
	CallTool(context.Context, *mcp.CallToolParams) (*mcp.CallToolResult, error)
}
