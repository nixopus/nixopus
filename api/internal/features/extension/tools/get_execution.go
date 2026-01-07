package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	extension_service "github.com/raghavyuva/nixopus-api/internal/features/extension/service"
	extension_storage "github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// GetExecutionHandler returns the handler function for getting an execution
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func GetExecutionHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, GetExecutionInput) (*mcp.CallToolResult, GetExecutionOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetExecutionInput,
	) (*mcp.CallToolResult, GetExecutionOutput, error) {
		storage := extension_storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
		service := extension_service.NewExtensionService(store, ctx, l, &storage)

		execution, err := service.GetExecutionByID(input.ExecutionID)
		if err != nil {
			return nil, GetExecutionOutput{}, err
		}

		// Convert shared types to MCP types to avoid circular references
		mcpExecution := convertToMCPExtensionExecution(*execution)

		return nil, GetExecutionOutput{
			Execution: mcpExecution,
		}, nil
	}
}
