package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// PruneBuildCacheHandler returns the handler function for pruning build cache
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func PruneBuildCacheHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, PruneBuildCacheInput) (*mcp.CallToolResult, PruneBuildCacheOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input PruneBuildCacheInput,
	) (*mcp.CallToolResult, PruneBuildCacheOutput, error) {
		// Build options with defaults
		opts := service.PruneBuildCacheOptions{
			All: false,
		}

		// Override with provided values
		if input.All != nil {
			opts.All = *input.All
		}

		response, err := service.PruneBuildCache(dockerService, l, opts)
		if err != nil {
			return nil, PruneBuildCacheOutput{}, err
		}

		return nil, PruneBuildCacheOutput{
			Response: response,
		}, nil
	}
}
