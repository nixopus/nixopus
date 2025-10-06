package tasks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// HandleReDeploy clones source, builds image using redeploy flags, and atomically updates the container
func (s *TaskService) HandleReDeploy(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	taskCtx := s.NewTaskContext(TaskPayload)

	taskCtx.LogAndUpdateStatus("Starting redeploy process", shared_types.Cloning)

	repoPath, err := s.Clone(CloneConfig{
		TaskPayload:    TaskPayload,
		DeploymentType: string(shared_types.DeploymentTypeReDeploy),
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
		Force:             TaskPayload.UpdateOptions.Force,
		ForceWithoutCache: TaskPayload.UpdateOptions.ForceWithoutCache,
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
	taskCtx.LogAndUpdateStatus("Redeploy completed successfully", shared_types.Deployed)

	client := GetCaddyClient()
	port, err := strconv.Atoi(containerResult.AvailablePort)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to convert port to int: "+err.Error(), shared_types.Failed)
		return err
	}
	upstreamHost := config.AppConfig.SSH.Host

	err = client.AddDomainWithAutoTLS(TaskPayload.Application.Domain, upstreamHost, port, caddygo.DomainOptions{})
	if err != nil {
		fmt.Println("Failed to add domain: ", err)
		taskCtx.LogAndUpdateStatus("Failed to add domain: "+err.Error(), shared_types.Failed)
		return err
	}
	client.Reload()

	return nil
}
