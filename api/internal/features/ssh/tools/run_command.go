package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// RunCommandHandler returns the handler function for running a command on a remote server via SSH
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func RunCommandHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, RunCommandInput) (*mcp.CallToolResult, RunCommandOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input RunCommandInput,
	) (*mcp.CallToolResult, RunCommandOutput, error) {
		manager := ssh.GetSSHManager()

		var output string
		var err error

		if input.ClientID != "" {
			output, err = manager.RunCommandWithID(input.ClientID, input.Command)
		} else {
			output, err = manager.RunCommand(input.Command)
		}

		if err != nil {
			// Return output even if there's an error (command might have produced output before failing)
			// Exit code is typically non zero when err != nil, but we can't determine exact exit code
			// from the current SSH implementation, so we'll use -1 to indicate error
			return nil, RunCommandOutput{
				Output:   output,
				ExitCode: -1,
			}, err
		}

		return nil, RunCommandOutput{
			Output:   output,
			ExitCode: 0,
		}, nil
	}
}
