package tasks

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

var (
	// ErrDeploymentCancelled is returned when a deployment is cancelled
	ErrDeploymentCancelled = errors.New("deployment was cancelled")
)

// CancellationRegistry tracks ongoing deployments and their cancellation functions
type CancellationRegistry struct {
	mu      sync.RWMutex
	cancels map[uuid.UUID]context.CancelFunc
}

var globalRegistry = &CancellationRegistry{
	cancels: make(map[uuid.UUID]context.CancelFunc),
}

// GetCancellationRegistry returns the global cancellation registry
func GetCancellationRegistry() *CancellationRegistry {
	return globalRegistry
}

// Register registers a deployment's cancel function
func (r *CancellationRegistry) Register(deploymentID uuid.UUID, cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancels[deploymentID] = cancel
}

// Unregister removes a deployment's cancel function (when task completes)
func (r *CancellationRegistry) Unregister(deploymentID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cancels, deploymentID)
}

// Cancel cancels a deployment if it's running
// Returns true if the deployment was found and cancelled, false otherwise
func (r *CancellationRegistry) Cancel(deploymentID uuid.UUID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cancel, exists := r.cancels[deploymentID]; exists {
		cancel()
		delete(r.cancels, deploymentID)
		return true
	}
	return false
}

// IsRunning checks if a deployment is currently running
func (r *CancellationRegistry) IsRunning(deploymentID uuid.UUID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.cancels[deploymentID]
	return exists
}

// CreateCancellableContext creates a cancellable context and registers it
func (r *CancellationRegistry) CreateCancellableContext(parentCtx context.Context, deploymentID uuid.UUID) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parentCtx)
	r.Register(deploymentID, cancel)

	// Create a wrapped cancel that also unregisters
	wrappedCancel := func() {
		cancel()
		r.Unregister(deploymentID)
	}

	return ctx, wrappedCancel
}

// CheckCancellation checks if the context is cancelled and returns an error if so
func CheckCancellation(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ErrDeploymentCancelled
	default:
		return nil
	}
}
