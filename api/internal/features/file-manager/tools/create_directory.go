package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	file_manager_service "github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	file_manager_types "github.com/raghavyuva/nixopus-api/internal/features/file-manager/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// CreateDirectoryHandler returns the handler function for creating a directory
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func CreateDirectoryHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, CreateDirectoryInput) (*mcp.CallToolResult, CreateDirectoryOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input CreateDirectoryInput,
	) (*mcp.CallToolResult, CreateDirectoryOutput, error) {
		fileManagerService := file_manager_service.NewFileManagerService(ctx, l)
		err := fileManagerService.CreateDirectory(input.Path)
		if err != nil {
			return nil, CreateDirectoryOutput{}, err
		}

		return nil, CreateDirectoryOutput{
			Response: file_manager_types.MessageResponse{
				Status:  "success",
				Message: "Directory created successfully",
			},
		}, nil
	}
}
