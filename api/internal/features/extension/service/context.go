package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type RunContext struct {
	Exec  *types.ExtensionExecution
	Spec  types.ExtensionSpec
	Vars  map[string]interface{}
	SSH   *ssh.SSH
	Steps []types.ExecutionStep
	// Compensations is a stack of revert functions for previously completed steps
	compensations []func()
	rolledBack    bool
}

func NewRunContext(exec *types.ExtensionExecution, spec types.ExtensionSpec, vars map[string]interface{}, sshClient *ssh.SSH, steps []types.ExecutionStep) *RunContext {
	return &RunContext{
		Exec:          exec,
		Spec:          spec,
		Vars:          vars,
		SSH:           sshClient,
		Steps:         steps,
		compensations: make([]func(), 0, 8),
	}
}

// pushCompensation adds a revert function to the stack
func (c *RunContext) pushCompensation(fn func()) {
	if fn == nil {
		return
	}
	c.compensations = append(c.compensations, fn)
}

// rollback executes compensations in reverse order, ignoring errors
func (c *RunContext) rollback() {
	if c.rolledBack {
		return
	}
	for i := len(c.compensations) - 1; i >= 0; i-- {
		c.compensations[i]()
	}
	c.compensations = nil
	c.rolledBack = true
}
