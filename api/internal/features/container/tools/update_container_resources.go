package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// UpdateContainerResourcesHandler returns the handler function for updating container resources
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func UpdateContainerResourcesHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, UpdateContainerResourcesInput) (*mcp.CallToolResult, UpdateContainerResourcesOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input UpdateContainerResourcesInput,
	) (*mcp.CallToolResult, UpdateContainerResourcesOutput, error) {
		// Build options with defaults (0 means unlimited/not set)
		opts := service.UpdateContainerResourcesOptions{
			ContainerID: input.ID,
			Memory:      0,
			MemorySwap:  0,
			CPUShares:   0,
		}

		// Override with provided values
		if input.Memory != nil {
			opts.Memory = *input.Memory
		}
		if input.MemorySwap != nil {
			opts.MemorySwap = *input.MemorySwap
		}
		if input.CPUShares != nil {
			opts.CPUShares = *input.CPUShares
		}

		response, err := service.UpdateContainerResources(dockerService, l, opts)
		if err != nil {
			return nil, UpdateContainerResourcesOutput{}, err
		}

		return nil, UpdateContainerResourcesOutput{
			Response: response,
		}, nil
	}
}
