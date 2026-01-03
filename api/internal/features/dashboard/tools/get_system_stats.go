package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/dashboard"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// GetSystemStatsHandler returns the handler function for getting system stats
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func GetSystemStatsHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, GetSystemStatsInput) (*mcp.CallToolResult, GetSystemStatsOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetSystemStatsInput,
	) (*mcp.CallToolResult, GetSystemStatsOutput, error) {
		stats, err := dashboard.CollectSystemStats(l, dashboard.GetSystemStatsOptions{})
		if err != nil {
			return nil, GetSystemStatsOutput{}, err
		}

		return nil, GetSystemStatsOutput{
			Stats: stats,
		}, nil
	}
}
