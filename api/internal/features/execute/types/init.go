package types

import "errors"

// ExecuteRequest represents a command execution request.
type ExecuteRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// ExecuteResponse represents the response from command execution.
type ExecuteResponse struct {
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exit_code"`
}

// AllowedCommands is the whitelist of commands that can be executed via the API.
var AllowedCommands = map[string]bool{
	"trail": true,
	"lxd":   true,
}

// Domain errors for execute feature.
var (
	ErrCommandRequired    = errors.New("command is required")
	ErrCommandNotAllowed  = errors.New("command not allowed. permitted commands: trail, lxd")
	ErrInvalidRequestType = errors.New("invalid request type")
	ErrExecutionFailed    = errors.New("command execution failed")
)
