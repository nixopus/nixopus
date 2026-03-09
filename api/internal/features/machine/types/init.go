package types

import (
	"github.com/nixopus/nixopus/api/internal/features/dashboard"
)

type SystemStatsResponse struct {
	Status  string                `json:"status"`
	Message string                `json:"message"`
	Data    dashboard.SystemStats `json:"data"`
}

type HostExecRequest struct {
	Command string `json:"command" validate:"required"`
}

type HostExecResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Data    HostExecData `json:"data"`
}

type HostExecData struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}
