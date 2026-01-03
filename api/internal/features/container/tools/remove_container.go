package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// RemoveContainerHandler returns the handler function for removing a container
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func RemoveContainerHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, RemoveContainerInput) (*mcp.CallToolResult, RemoveContainerOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input RemoveContainerInput,
	) (*mcp.CallToolResult, RemoveContainerOutput, error) {
		// Build options with defaults
		opts := service.RemoveContainerOptions{
			ContainerID: input.ID,
			Force:       true, // Default to force removal
		}

		// Override with provided values
		if input.Force != nil {
			opts.Force = *input.Force
		}

		response, err := service.RemoveContainer(dockerService, l, opts)
		if err != nil {
			return nil, RemoveContainerOutput{}, err
		}

		return nil, RemoveContainerOutput{
			Response: response,
		}, nil
	}
}
