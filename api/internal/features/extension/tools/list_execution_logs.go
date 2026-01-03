package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	extension_service "github.com/raghavyuva/nixopus-api/internal/features/extension/service"
	extension_storage "github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// ListExecutionLogsHandler returns the handler function for listing execution logs
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func ListExecutionLogsHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, ListExecutionLogsInput) (*mcp.CallToolResult, ListExecutionLogsOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input ListExecutionLogsInput,
	) (*mcp.CallToolResult, ListExecutionLogsOutput, error) {
		storage := extension_storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
		service := extension_service.NewExtensionService(store, ctx, l, &storage)

		afterSeq := int64(0)
		if input.AfterSeq != nil {
			afterSeq = *input.AfterSeq
		}

		limit := 200
		if input.Limit != nil {
			limit = *input.Limit
		}

		logs, execStatus, err := service.ListExecutionLogs(input.ExecutionID, afterSeq, limit)
		if err != nil {
			return nil, ListExecutionLogsOutput{}, err
		}

		nextAfter := afterSeq
		if len(logs) > 0 {
			nextAfter = logs[len(logs)-1].Sequence
		}

		return nil, ListExecutionLogsOutput{
			Logs:            logs,
			NextAfter:       nextAfter,
			ExecutionStatus: execStatus,
		}, nil
	}
}
