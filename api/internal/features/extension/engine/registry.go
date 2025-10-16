package engine

import (
	"sync"

	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// Module defines a pluggable executor for a step type.
type Module interface {
	// Type returns the step type this module handles
	Type() string
	// Execute runs the step and returns output or an error.
	Execute(sshClient *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error)
}

var (
	regMu   sync.RWMutex
	modules = map[string]Module{}
)

// RegisterModule registers a module by its Type().
func RegisterModule(m Module) {
	if m == nil {
		return
	}
	regMu.Lock()
	modules[m.Type()] = m
	regMu.Unlock()
}

// GetModule fetches a module by type.
func GetModule(stepType string) Module {
	regMu.RLock()
	m := modules[stepType]
	regMu.RUnlock()
	return m
}
