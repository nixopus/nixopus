package tasks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// RestartDeployment enqueues a restart task for an application deployment
func (t *TaskService) RestartDeployment(request *types.RestartDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID) error {
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

	payload, err := ctxTask.PrepareRestartContext()
	if err != nil {
		return err
	}

	payload.CorrelationID = uuid.NewString()

	return RestartQueue.Add(TaskRestart.WithArgs(context.Background(), payload))
}

// HandleRestart routes restart based on the application's BuildPack type
func (s *TaskService) HandleRestart(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	switch TaskPayload.Application.BuildPack {
	case shared_types.DockerFile:
		return s.HandleRestartDockerfileDeployment(ctx, TaskPayload)
	case shared_types.DockerCompose:
		return s.HandleRestartDockerComposeDeployment(ctx, TaskPayload)
	case shared_types.Static:
		return s.HandleRestartStaticDeployment(ctx, TaskPayload)
	default:
		return types.ErrInvalidBuildPack
	}
}

// HandleRestartDockerfileDeployment restarts currently running swarm service for the application
func (s *TaskService) HandleRestartDockerfileDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	taskCtx := s.NewTaskContext(TaskPayload)

	taskCtx.LogAndUpdateStatus("Restarting application service", shared_types.Deploying)

	ctx = context.WithValue(ctx, shared_types.OrganizationIDKey, TaskPayload.Application.OrganizationID.String())

	dockerService, err := s.getDockerService(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get docker service: "+err.Error(), shared_types.Failed)
		return err
	}

	// Find the existing service
	existingService, err := s.getExistingService(ctx, TaskPayload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to find service: "+err.Error(), shared_types.Failed)
		return err
	}

	if existingService == nil {
		taskCtx.LogAndUpdateStatus("No running service found for application", shared_types.Failed)
		return types.ErrContainerNotRunning
	}

	taskCtx.AddLog("Restarting service " + existingService.ID)

	currentService, err := dockerService.GetServiceByID(existingService.ID)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get service details: "+err.Error(), shared_types.Failed)
		return err
	}

	if currentService.ID == "" {
		taskCtx.LogAndUpdateStatus("Service not found", shared_types.Failed)
		return types.ErrContainerNotRunning
	}

	err = dockerService.UpdateService(existingService.ID, currentService.Spec, "")
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to restart service: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.LogAndUpdateStatus("Application service restarted", shared_types.Running)
	return nil
}

// HandleRestartDockerComposeDeployment handles restart of a Docker Compose application
func (s *TaskService) HandleRestartDockerComposeDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return s.deployDockerCompose(ctx, TaskPayload, string(shared_types.DeploymentTypeRestart))
}

// HandleRestartStaticDeployment handles restart of a static application
func (s *TaskService) HandleRestartStaticDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	// TODO: Implement static restart
	return fmt.Errorf("static restart not yet implemented")
}
