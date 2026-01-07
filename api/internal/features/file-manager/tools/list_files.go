package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	file_manager_service "github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	file_manager_types "github.com/raghavyuva/nixopus-api/internal/features/file-manager/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// ListFilesHandler returns the handler function for listing files
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func ListFilesHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, ListFilesInput) (*mcp.CallToolResult, ListFilesOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input ListFilesInput,
	) (*mcp.CallToolResult, ListFilesOutput, error) {
		fileManagerService := file_manager_service.NewFileManagerService(ctx, l)
		files, err := fileManagerService.ListFiles(input.Path)
		if err != nil {
			return nil, ListFilesOutput{}, err
		}

		return nil, ListFilesOutput{
			Response: file_manager_types.ListFilesResponse{
				Status:  "success",
				Message: "Files fetched successfully",
				Data:    files,
			},
		}, nil
	}
}
