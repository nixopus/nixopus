package tasks

import (
    "context"

    "github.com/docker/docker/api/types/container"
    "github.com/google/uuid"
    "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
    shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// RestartDeployment enqueues a restart task for an application deployment
func (t *TaskService) RestartDeployment(request *types.RestartDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID) error {
    // Load target deployment and owning application
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

// HandleRestart restarts currently running containers for the application and updates status/logs
func (s *TaskService) HandleRestart(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
    taskCtx := s.NewTaskContext(TaskPayload)

    taskCtx.LogAndUpdateStatus("Restarting application containers", shared_types.Deploying)

    currentContainers, err := s.getRunningContainers(TaskPayload, taskCtx)
    if err != nil {
        taskCtx.LogAndUpdateStatus("Failed to list running containers: "+err.Error(), shared_types.Failed)
        return err
    }

    if len(currentContainers) == 0 {
        taskCtx.LogAndUpdateStatus("No running containers found for application", shared_types.Failed)
        return types.ErrContainerNotRunning
    }

    for _, ctr := range currentContainers {
        taskCtx.AddLog("Restarting container " + ctr.ID)
        if err := s.DockerRepo.RestartContainer(ctr.ID, container.StopOptions{}); err != nil {
            taskCtx.LogAndUpdateStatus("Failed to restart container: "+err.Error(), shared_types.Failed)
            return err
        }
    }

    taskCtx.LogAndUpdateStatus("Application containers restarted", shared_types.Running)
    return nil
}
