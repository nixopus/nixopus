package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// ListContainersHandler returns the handler function for listing containers
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func ListContainersHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, ListContainersInput) (*mcp.CallToolResult, ListContainersOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input ListContainersInput,
	) (*mcp.CallToolResult, ListContainersOutput, error) {
		// Build params with defaults
		params := container_types.ContainerListParams{
			Page:      1,
			PageSize:  10,
			SortBy:    "name",
			SortOrder: "asc",
		}

		// Override with provided values
		if input.Page != nil && *input.Page > 0 {
			params.Page = *input.Page
		}
		if input.PageSize != nil && *input.PageSize > 0 {
			params.PageSize = *input.PageSize
		}
		if input.SortBy != "" {
			params.SortBy = input.SortBy
		}
		if input.SortOrder != "" {
			params.SortOrder = input.SortOrder
		}
		if input.Search != "" {
			params.Search = input.Search
		}
		if input.Status != "" {
			params.Status = input.Status
		}
		if input.Name != "" {
			params.Name = input.Name
		}
		if input.Image != "" {
			params.Image = input.Image
		}

		response, err := service.ListContainers(dockerService, l, params)
		if err != nil {
			return nil, ListContainersOutput{}, err
		}

		return nil, ListContainersOutput{
			Response: response,
		}, nil
	}
}
