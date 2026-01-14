package tasks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// RollbackDeployment enqueues a rollback task to rebuild and deploy a previous commit
func (t *TaskService) RollbackDeployment(request *types.RollbackDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID) error {
	dep, err := t.Storage.GetApplicationDeploymentById(request.ID.String())
	if err != nil {
		return err
	}

	app, err := t.Storage.GetApplicationById(dep.ApplicationID.String(), organizationID)
	if err != nil {
		return err
	}

	ctxTask := ContextTask{
		TaskService:    t,
		ContextConfig:  request,
		UserId:         userID,
		OrganizationId: organizationID,
		Application:    &app,
	}

	payload, err := ctxTask.PrepareRollbackContext()
	if err != nil {
		return err
	}

	payload.CorrelationID = uuid.NewString()

	return RollbackQueue.Add(TaskRollback.WithArgs(context.Background(), payload))
}

// HandleRollback uses Docker Swarm's native rollback capability for instant rollback
func (s *TaskService) HandleRollback(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	taskCtx := s.NewTaskContext(TaskPayload)

	taskCtx.LogAndUpdateStatus("Starting native swarm rollback", shared_types.Deploying)

	serviceID := TaskPayload.ApplicationDeployment.ContainerID
	if serviceID == "" {
		taskCtx.LogAndUpdateStatus("No service ID found in deployment record", shared_types.Failed)
		return types.ErrContainerNotRunning
	}

	taskCtx.AddLog("Rolling back service " + serviceID + " using Docker Swarm native rollback")

	err := s.DockerRepo.RollbackService(serviceID)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to rollback service: "+err.Error(), shared_types.Failed)
		return err
	}

	// Wait for rollback to complete
	time.Sleep(time.Second * 5)

	serviceInfo, err := s.DockerRepo.GetServiceByID(serviceID)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get service info after rollback: "+err.Error(), shared_types.Failed)
		return err
	}

	// Check service's health
	if serviceInfo.Spec.Mode.Replicated != nil && serviceInfo.Spec.Mode.Replicated.Replicas != nil {
		running, _, err := s.DockerRepo.GetServiceHealth(serviceInfo)
		if err != nil || running < int(*serviceInfo.Spec.Mode.Replicated.Replicas) {
			taskCtx.LogAndUpdateStatus("Service health check failed after rollback", shared_types.Failed)
			return types.ErrFailedToUpdateContainer
		}
	}

	taskCtx.AddLog("Service rolled back successfully using native swarm rollback")
	taskCtx.LogAndUpdateStatus("Rollback completed successfully", shared_types.Deployed)

	return nil
}
