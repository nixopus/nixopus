package tasks

import (
	"context"

	"github.com/docker/docker/api/types/swarm"
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

// HandleRestart restarts currently running swarm service for the application and updates status/logs
func (s *TaskService) HandleRestart(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	taskCtx := s.NewTaskContext(TaskPayload)

	taskCtx.LogAndUpdateStatus("Restarting application service", shared_types.Deploying)

	// Find the existing service
	existingService, err := s.getExistingService(TaskPayload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to find service: "+err.Error(), shared_types.Failed)
		return err
	}

	if existingService == nil {
		taskCtx.LogAndUpdateStatus("No running service found for application", shared_types.Failed)
		return types.ErrContainerNotRunning
	}

	taskCtx.AddLog("Restarting service " + existingService.ID)

	// Get current service spec
	services, err := s.DockerRepo.GetClusterServices()
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get service details: "+err.Error(), shared_types.Failed)
		return err
	}

	var currentService swarm.Service
	for _, service := range services {
		if service.ID == existingService.ID {
			currentService = service
			break
		}
	}

	if currentService.ID == "" {
		taskCtx.LogAndUpdateStatus("Service not found", shared_types.Failed)
		return types.ErrContainerNotRunning
	}

	// Note : Restart service by updating it with the same spec will restart the service so we don't need to specifically restart the services
	err = s.DockerRepo.UpdateService(existingService.ID, currentService.Spec, "")
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to restart service: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.LogAndUpdateStatus("Application service restarted", shared_types.Running)
	return nil
}
