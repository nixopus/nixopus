package tasks

import (
	"context"
	"fmt"
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

// HandleRollback routes rollback based on the application's BuildPack type
func (s *TaskService) HandleRollback(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	switch TaskPayload.Application.BuildPack {
	case shared_types.DockerFile:
		return s.HandleRollbackDockerfileDeployment(ctx, TaskPayload)
	case shared_types.DockerCompose:
		return s.HandleRollbackDockerComposeDeployment(ctx, TaskPayload)
	case shared_types.Static:
		return s.HandleRollbackStaticDeployment(ctx, TaskPayload)
	default:
		return types.ErrInvalidBuildPack
	}
}

// HandleRollbackDockerfileDeployment uses Docker Swarm's native rollback capability for instant rollback
func (s *TaskService) HandleRollbackDockerfileDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
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

// HandleRollbackDockerComposeDeployment handles rollback of a Docker Compose application
func (s *TaskService) HandleRollbackDockerComposeDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return s.deployDockerCompose(ctx, TaskPayload, string(shared_types.DeploymentTypeRollback))
}

// HandleRollbackStaticDeployment handles rollback of a static application
func (s *TaskService) HandleRollbackStaticDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	// TODO: Implement static rollback
	return fmt.Errorf("static rollback not yet implemented")
}
