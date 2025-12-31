package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// DeployService handles read operations and validation for deployments.
// Deployment creation, update, rollback, restart, and delete operations
// are handled by TaskService in the tasks package.
type DeployService struct {
	storage storage.DeployRepository
	Ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewDeployService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, deploy_repo storage.DeployRepository) *DeployService {
	return &DeployService{
		storage: deploy_repo,
		store:   store,
		Ctx:     ctx,
		logger:  logger,
	}
}

// CancelDeployment cancels an ongoing deployment by updating its status to cancelled.
// It verifies the deployment is in progress before cancelling.
func (s *DeployService) CancelDeployment(deploymentID uuid.UUID, organizationID uuid.UUID) error {
	// Get the deployment to verify it exists and get the current status
	deployment, err := s.storage.GetApplicationDeploymentById(deploymentID.String())
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get deployment", err.Error())
		return err
	}

	// Check if deployment status exists
	if deployment.Status == nil {
		s.logger.Log(logger.Error, "Deployment status not found", deploymentID.String())
		return types.ErrDeploymentNotInProgress
	}

	// Check if deployment is already cancelled
	if deployment.Status.Status == shared_types.Cancelled {
		return types.ErrDeploymentAlreadyCancelled
	}

	// Check if deployment is in progress (can be cancelled)
	if !isInProgressStatus(deployment.Status.Status) {
		s.logger.Log(logger.Info, "Deployment is not in progress, cannot cancel", "status: "+string(deployment.Status.Status))
		return types.ErrDeploymentNotInProgress
	}

	// Try to cancel the running task if it exists
	registry := tasks.GetCancellationRegistry()
	if registry.Cancel(deploymentID) {
		s.logger.Log(logger.Info, "Running task cancelled via context", deploymentID.String())
	} else {
		// Task might not be running yet (queued) or already finished
		// Still update the status in database
		s.logger.Log(logger.Info, "No running task found, updating status directly", deploymentID.String())
	}

	// Update the deployment status to cancelled in database
	err = s.storage.CancelDeployment(deploymentID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to cancel deployment", err.Error())
		return err
	}

	s.logger.Log(logger.Info, "Deployment cancelled successfully", deploymentID.String())
	return nil
}

// isInProgressStatus checks if the given status represents an in-progress deployment.
func isInProgressStatus(status shared_types.Status) bool {
	switch status {
	case shared_types.Started, shared_types.Cloning, shared_types.Building, shared_types.Deploying:
		return true
	default:
		return false
	}
}
