package types

import "errors"

type MachineState struct {
	Active    bool   `json:"active"`
	State     string `json:"state"`
	PID       int    `json:"pid"`
	UptimeSec int64  `json:"uptime_sec"`
}

type MachineStateResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Data    *MachineState `json:"data"`
}

type MachineActionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	ErrMachineNotProvisioned   = errors.New("no provisioned machine found")
	ErrMachineOperationTimeout = errors.New("machine operation timed out")
	ErrMachineOperationLocked  = errors.New("another operation is in progress")
	ErrMachineNotRunning       = errors.New("machine is not running")
	ErrMachineAlreadyPaused    = errors.New("machine is already paused")
	ErrMachineNotPaused        = errors.New("machine is not paused")
)
