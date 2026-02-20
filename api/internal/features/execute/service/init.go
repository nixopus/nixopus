package service

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/execute/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// ExecuteService handles business logic for command execution.
type ExecuteService struct {
	logger      logger.Logger
	abyssBinary string
}

// NewExecuteService creates a new ExecuteService instance.
func NewExecuteService(l logger.Logger) *ExecuteService {
	return &ExecuteService{
		logger:      l,
		abyssBinary: "abyss",
	}
}

// ExecuteCommand executes a command via the abyss CLI.
//
// Parameters:
//   - userID: the UUID of the requesting user (for logging)
//   - req: the execute request containing command and args
//
// Returns:
//   - *types.ExecuteResponse: the execution result
//   - error: execution error if command fails
func (s *ExecuteService) ExecuteCommand(userID string, req types.ExecuteRequest) (*types.ExecuteResponse, error) {
	s.logger.Log(logger.Info, fmt.Sprintf("User %s executing command: %s %v", userID, req.Command, req.Args), userID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := append([]string{req.Command}, req.Args...)

	cmd := exec.CommandContext(ctx, s.abyssBinary, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	response := &types.ExecuteResponse{
		Output:   stdout.String(),
		Error:    stderr.String(),
		ExitCode: exitCode,
	}

	if ctx.Err() == context.DeadlineExceeded {
		response.Error = "command execution timed out"
		response.ExitCode = 124
	}

	return response, nil
}
