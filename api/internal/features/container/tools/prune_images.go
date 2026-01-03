package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// PruneImagesHandler returns the handler function for pruning images
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func PruneImagesHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, PruneImagesInput) (*mcp.CallToolResult, PruneImagesOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input PruneImagesInput,
	) (*mcp.CallToolResult, PruneImagesOutput, error) {
		// Build options with defaults
		opts := service.PruneImagesOptions{
			Until:    input.Until,
			Label:    input.Label,
			Dangling: false,
		}

		// Override with provided values
		if input.Dangling != nil {
			opts.Dangling = *input.Dangling
		}

		response, err := service.PruneImages(dockerService, l, opts)
		if err != nil {
			return nil, PruneImagesOutput{}, err
		}

		return nil, PruneImagesOutput{
			Response: response,
		}, nil
	}
}
