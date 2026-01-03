package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// StartContainerHandler returns the handler function for starting a container
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func StartContainerHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, StartContainerInput) (*mcp.CallToolResult, StartContainerOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input StartContainerInput,
	) (*mcp.CallToolResult, StartContainerOutput, error) {
		opts := service.StartContainerOptions{
			ContainerID: input.ID,
		}

		response, err := service.StartContainer(dockerService, l, opts)
		if err != nil {
			return nil, StartContainerOutput{}, err
		}

		return nil, StartContainerOutput{
			Response: response,
		}, nil
	}
}
