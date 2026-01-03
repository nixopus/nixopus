package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// StopContainerHandler returns the handler function for stopping a container
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func StopContainerHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, StopContainerInput) (*mcp.CallToolResult, StopContainerOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input StopContainerInput,
	) (*mcp.CallToolResult, StopContainerOutput, error) {
		// Build options with defaults
		opts := service.StopContainerOptions{
			ContainerID: input.ID,
			Timeout:     nil, // Default timeout handled by Docker
		}

		// Override with provided values
		if input.Timeout != nil {
			opts.Timeout = input.Timeout
		}

		response, err := service.StopContainer(dockerService, l, opts)
		if err != nil {
			return nil, StopContainerOutput{}, err
		}

		return nil, StopContainerOutput{
			Response: response,
		}, nil
	}
}
