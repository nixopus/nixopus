package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	mcp_middleware "github.com/raghavyuva/nixopus-api/internal/mcp/middleware"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// GetContainerLogsHandler returns the handler function for getting container logs
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func GetContainerLogsHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dockerService *docker.DockerService,
) func(context.Context, *mcp.CallToolRequest, GetContainerLogsInput) (*mcp.CallToolResult, GetContainerLogsOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetContainerLogsInput,
	) (*mcp.CallToolResult, GetContainerLogsOutput, error) {
		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero GetContainerLogsOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}

		// Handle optional pointer fields
		tail := 0
		if input.Tail != nil {
			tail = *input.Tail
		}
		var since, until string
		if input.Since != nil {
			since = *input.Since
		}
		if input.Until != nil {
			until = *input.Until
		}

		decodedLogs, err := service.GetContainerLogs(
			toolCtx,
			store,
			dockerService,
			l,
			service.ContainerLogsOptions{
				ContainerID:    input.ID,
				OrganizationID: orgID,
				Follow:         input.Follow,
				Tail:           tail,
				Since:          since,
				Until:          until,
				Stdout:         input.Stdout,
				Stderr:         input.Stderr,
			},
		)
		if err != nil {
			return nil, GetContainerLogsOutput{}, err
		}

		return nil, GetContainerLogsOutput{
			Logs: decodedLogs,
		}, nil
	}
}
