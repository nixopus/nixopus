package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	file_manager_service "github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	file_manager_types "github.com/raghavyuva/nixopus-api/internal/features/file-manager/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// CopyDirectoryHandler returns the handler function for copying a file or directory
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func CopyDirectoryHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, CopyDirectoryInput) (*mcp.CallToolResult, CopyDirectoryOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input CopyDirectoryInput,
	) (*mcp.CallToolResult, CopyDirectoryOutput, error) {
		fileManagerService := file_manager_service.NewFileManagerService(ctx, l)
		err := fileManagerService.CopyDirectory(input.FromPath, input.ToPath)
		if err != nil {
			return nil, CopyDirectoryOutput{}, err
		}

		return nil, CopyDirectoryOutput{
			Response: file_manager_types.MessageResponse{
				Status:  "success",
				Message: "File or directory copied successfully",
			},
		}, nil
	}
}
