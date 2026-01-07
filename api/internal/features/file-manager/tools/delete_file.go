package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	file_manager_service "github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	file_manager_types "github.com/raghavyuva/nixopus-api/internal/features/file-manager/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// DeleteFileHandler returns the handler function for deleting a file or directory
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func DeleteFileHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, DeleteFileInput) (*mcp.CallToolResult, DeleteFileOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input DeleteFileInput,
	) (*mcp.CallToolResult, DeleteFileOutput, error) {
		fileManagerService := file_manager_service.NewFileManagerService(ctx, l)
		err := fileManagerService.DeleteFile(input.Path)
		if err != nil {
			return nil, DeleteFileOutput{}, err
		}

		return nil, DeleteFileOutput{
			Response: file_manager_types.MessageResponse{
				Status:  "success",
				Message: "File or directory deleted successfully",
			},
		}, nil
	}
}
