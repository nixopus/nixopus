package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	extension_service "github.com/raghavyuva/nixopus-api/internal/features/extension/service"
	extension_storage "github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// GetExtensionHandler returns the handler function for getting a single extension
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func GetExtensionHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, GetExtensionInput) (*mcp.CallToolResult, GetExtensionOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetExtensionInput,
	) (*mcp.CallToolResult, GetExtensionOutput, error) {
		storage := extension_storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
		service := extension_service.NewExtensionService(store, ctx, l, &storage)

		extension, err := service.GetExtension(input.ID)
		if err != nil {
			return nil, GetExtensionOutput{}, err
		}

		// Convert shared types to MCP types to avoid circular references
		mcpExtension := convertToMCPExtension(*extension)

		return nil, GetExtensionOutput{
			Extension: mcpExtension,
		}, nil
	}
}
