package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// GetContainerHandler returns the handler function for getting container details
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func GetContainerHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, GetContainerInput) (*mcp.CallToolResult, GetContainerOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetContainerInput,
	) (*mcp.CallToolResult, GetContainerOutput, error) {
		containerData, err := service.GetContainer(dockerService, l, input.ID)
		if err != nil {
			return nil, GetContainerOutput{}, err
		}

		return nil, GetContainerOutput{
			Container: containerData,
		}, nil
	}
}
