package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// RestartContainerHandler returns the handler function for restarting a container
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func RestartContainerHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, RestartContainerInput) (*mcp.CallToolResult, RestartContainerOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input RestartContainerInput,
	) (*mcp.CallToolResult, RestartContainerOutput, error) {
		// Build options with defaults
		opts := service.RestartContainerOptions{
			ContainerID: input.ID,
			Timeout:     nil, // Default timeout handled by Docker
		}

		// Override with provided values
		if input.Timeout != nil {
			opts.Timeout = input.Timeout
		}

		response, err := service.RestartContainer(dockerService, l, opts)
		if err != nil {
			return nil, RestartContainerOutput{}, err
		}

		return nil, RestartContainerOutput{
			Response: response,
		}, nil
	}
}
