package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// ListImagesHandler returns the handler function for listing images
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func ListImagesHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, ListImagesInput) (*mcp.CallToolResult, ListImagesOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input ListImagesInput,
	) (*mcp.CallToolResult, ListImagesOutput, error) {
		// Build options with defaults
		opts := service.ListImagesOptions{
			All:         false,
			ContainerID: input.ContainerID,
			ImagePrefix: input.ImagePrefix,
		}

		// Override with provided values
		if input.All != nil {
			opts.All = *input.All
		}

		response, err := service.ListImages(dockerService, l, opts)
		if err != nil {
			return nil, ListImagesOutput{}, err
		}

		return nil, ListImagesOutput{
			Response: response,
		}, nil
	}
}
