package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	extension_service "github.com/raghavyuva/nixopus-api/internal/features/extension/service"
	extension_storage "github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// RunExtensionHandler returns the handler function for running an extension
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func RunExtensionHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, RunExtensionInput) (*mcp.CallToolResult, RunExtensionOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input RunExtensionInput,
	) (*mcp.CallToolResult, RunExtensionOutput, error) {
		storage := extension_storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
		service := extension_service.NewExtensionService(store, ctx, l, &storage)

		variables := input.Variables
		if variables == nil {
			variables = make(map[string]interface{})
		}

		execution, err := service.StartRun(input.ExtensionID, variables)
		if err != nil {
			return nil, RunExtensionOutput{}, err
		}

		return nil, RunExtensionOutput{
			Execution: *execution,
		}, nil
	}
}
