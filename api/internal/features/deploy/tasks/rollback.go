package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	s3store "github.com/raghavyuva/nixopus-api/internal/features/deploy/s3"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
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

// HandleRollback performs rollback via S3 image restore when available,
// falling back to Docker Swarm's native rollback otherwise.
func (s *TaskService) HandleRollback(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	taskCtx := s.NewTaskContext(TaskPayload)

	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, TaskPayload.Application.OrganizationID.String())

	if s3store.IsConfigured(config.AppConfig.S3) && TaskPayload.ApplicationDeployment.ImageS3Key != "" {
		return s.handleS3Rollback(orgCtx, TaskPayload, taskCtx)
	}

	return s.handleSwarmRollback(orgCtx, TaskPayload, taskCtx)
}

// handleS3Rollback loads the image from S3, tags it, and updates the swarm service.
func (s *TaskService) handleS3Rollback(ctx context.Context, TaskPayload shared_types.TaskPayload, taskCtx *TaskContext) error {
	taskCtx.LogAndUpdateStatus("Starting S3-based rollback", shared_types.Deploying)

	err := s.LoadImageFromS3(ctx, TaskPayload.ApplicationDeployment.ImageS3Key, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to load image from S3: "+err.Error(), shared_types.Failed)
		return err
	}

	commitTag := CommitImageTag(TaskPayload.Application.Name, TaskPayload.ApplicationDeployment.CommitHash)
	latestTag := fmt.Sprintf("%s:latest", TaskPayload.Application.Name)

	sshManager, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get SSH manager: "+err.Error(), shared_types.Failed)
		return err
	}

	clientConn, err := sshManager.Connect()
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to connect via SSH: "+err.Error(), shared_types.Failed)
		return err
	}

	session, err := clientConn.NewSession()
	if err != nil {
		clientConn.Close()
		taskCtx.LogAndUpdateStatus("Failed to create SSH session: "+err.Error(), shared_types.Failed)
		return err
	}

	tagCmd := fmt.Sprintf("docker tag %s %s", commitTag, latestTag)
	output, err := session.CombinedOutput(tagCmd)
	session.Close()
	clientConn.Close()

	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to tag image: "+err.Error()+" output: "+string(output), shared_types.Failed)
		return fmt.Errorf("docker tag failed: %w", err)
	}
	taskCtx.AddLog("Image tagged as " + latestTag)

	containerResult, err := s.AtomicUpdateContainer(ctx, TaskPayload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to update container: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.AddLog("Container updated successfully with container id " + containerResult.ContainerID)
	taskCtx.LogAndUpdateStatus("S3-based rollback completed successfully", shared_types.Deployed)
	return nil
}

// handleSwarmRollback uses Docker Swarm's native rollback capability.
func (s *TaskService) handleSwarmRollback(ctx context.Context, TaskPayload shared_types.TaskPayload, taskCtx *TaskContext) error {
	taskCtx.LogAndUpdateStatus("Starting native swarm rollback", shared_types.Deploying)

	serviceID := TaskPayload.ApplicationDeployment.ContainerID
	if serviceID == "" {
		taskCtx.LogAndUpdateStatus("No service ID found in deployment record", shared_types.Failed)
		return types.ErrContainerNotRunning
	}

	taskCtx.AddLog("Rolling back service " + serviceID + " using Docker Swarm native rollback")

	dockerService, err := s.getDockerService(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get docker service: "+err.Error(), shared_types.Failed)
		return err
	}

	err = dockerService.RollbackService(serviceID)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to rollback service: "+err.Error(), shared_types.Failed)
		return err
	}

	for i := 0; i < 15; i++ {
		time.Sleep(2 * time.Second)
		serviceInfo, getErr := dockerService.GetServiceByID(serviceID)
		if getErr != nil {
			continue
		}
		if serviceInfo.Spec.Mode.Replicated != nil && serviceInfo.Spec.Mode.Replicated.Replicas != nil {
			running, _, healthErr := dockerService.GetServiceHealth(serviceInfo)
			if healthErr == nil && running >= int(*serviceInfo.Spec.Mode.Replicated.Replicas) {
				taskCtx.AddLog("Service rolled back successfully using native swarm rollback")
				taskCtx.LogAndUpdateStatus("Rollback completed successfully", shared_types.Deployed)
				return nil
			}
		}
	}

	serviceInfo, err := dockerService.GetServiceByID(serviceID)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get service info after rollback: "+err.Error(), shared_types.Failed)
		return err
	}

	if serviceInfo.Spec.Mode.Replicated != nil && serviceInfo.Spec.Mode.Replicated.Replicas != nil {
		running, _, err := dockerService.GetServiceHealth(serviceInfo)
		if err != nil || running < int(*serviceInfo.Spec.Mode.Replicated.Replicas) {
			taskCtx.LogAndUpdateStatus("Service health check failed after rollback", shared_types.Failed)
			return types.ErrFailedToUpdateContainer
		}
	}

	taskCtx.AddLog("Service rolled back successfully using native swarm rollback")
	taskCtx.LogAndUpdateStatus("Rollback completed successfully", shared_types.Deployed)

	return nil
}
