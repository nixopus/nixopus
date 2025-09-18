package tasks

import (
    "context"

    "github.com/google/uuid"
    "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
    shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// RollbackDeployment enqueues a rollback task to rebuild and deploy a previous commit
func (t *TaskService) RollbackDeployment(request *types.RollbackDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID) error {
    // Find the target deployment and owning application
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

// HandleRollback clones checkout at target commit, rebuilds and atomically updates container
func (s *TaskService) HandleRollback(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
    taskCtx := s.NewTaskContext(TaskPayload)

    taskCtx.LogAndUpdateStatus("Starting rollback process", shared_types.Cloning)

    repoPath, err := s.Clone(CloneConfig{
        TaskPayload:    TaskPayload,
        DeploymentType: string(shared_types.DeploymentTypeRollback),
        TaskContext:    taskCtx,
    })
    if err != nil {
        taskCtx.LogAndUpdateStatus("Failed to clone repository: "+err.Error(), shared_types.Failed)
        return err
    }

    taskCtx.LogAndUpdateStatus("Repository cloned successfully", shared_types.Building)
    taskCtx.AddLog("Building image from Dockerfile " + repoPath + " for application " + TaskPayload.Application.Name)

    buildImageResult, err := s.BuildImage(BuildConfig{
        TaskPayload:       TaskPayload,
        ContextPath:       repoPath,
        Force:             false,
        ForceWithoutCache: false,
        TaskContext:       taskCtx,
    })
    if err != nil {
        taskCtx.LogAndUpdateStatus("Failed to build image: "+err.Error(), shared_types.Failed)
        return err
    }

    taskCtx.AddLog("Image built successfully: " + buildImageResult + " for application " + TaskPayload.Application.Name)
    taskCtx.UpdateStatus(shared_types.Deploying)

    containerResult, err := s.AtomicUpdateContainer(TaskPayload, taskCtx)
    if err != nil {
        taskCtx.LogAndUpdateStatus("Failed to update container: "+err.Error(), shared_types.Failed)
        return err
    }

    taskCtx.AddLog("Container updated successfully for application " + TaskPayload.Application.Name + " with container id " + containerResult.ContainerID)
    taskCtx.LogAndUpdateStatus("Rollback completed successfully", shared_types.Deployed)

    return nil
}
