package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	extension_service "github.com/raghavyuva/nixopus-api/internal/features/extension/service"
	extension_storage "github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// ListExtensionsHandler returns the handler function for listing extensions
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func ListExtensionsHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, ListExtensionsInput) (*mcp.CallToolResult, ListExtensionsOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input ListExtensionsInput,
	) (*mcp.CallToolResult, ListExtensionsOutput, error) {
		storage := extension_storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
		service := extension_service.NewExtensionService(store, ctx, l, &storage)

		params := shared_types.ExtensionListParams{}

		if input.Category != "" {
			cat := shared_types.ExtensionCategory(input.Category)
			params.Category = &cat
		}

		if input.Type != "" {
			et := shared_types.ExtensionType(input.Type)
			params.Type = &et
		}

		if input.Search != "" {
			params.Search = input.Search
		}

		if input.SortBy != "" {
			params.SortBy = shared_types.ExtensionSortField(input.SortBy)
		}

		if input.SortDir != "" {
			params.SortDir = shared_types.SortDirection(input.SortDir)
		}

		if input.Page != nil {
			params.Page = *input.Page
		}

		if input.PageSize != nil {
			params.PageSize = *input.PageSize
		}

		response, err := service.ListExtensions(params)
		if err != nil {
			return nil, ListExtensionsOutput{}, err
		}

		return nil, ListExtensionsOutput{
			Response: *response,
		}, nil
	}
}
