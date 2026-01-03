package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	extension_service "github.com/raghavyuva/nixopus-api/internal/features/extension/service"
	extension_storage "github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// CancelExecutionHandler returns the handler function for cancelling an execution
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func CancelExecutionHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, CancelExecutionInput) (*mcp.CallToolResult, CancelExecutionOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input CancelExecutionInput,
	) (*mcp.CallToolResult, CancelExecutionOutput, error) {
		storage := extension_storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
		service := extension_service.NewExtensionService(store, ctx, l, &storage)

		err := service.CancelExecution(input.ExecutionID)
		if err != nil {
			return nil, CancelExecutionOutput{}, err
		}

		return nil, CancelExecutionOutput{
			Message: "Execution cancelled successfully",
		}, nil
	}
}
