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

	if IsComposeDeployment(TaskPayload.Application) {
		return s.handleComposeReDeploy(ctx, TaskPayload, taskCtx)
	}

	return s.handleDockerfileReDeploy(ctx, TaskPayload, taskCtx)
}

// handleComposeReDeploy handles redeployment of docker-compose applications
func (s *TaskService) handleComposeReDeploy(ctx context.Context, TaskPayload shared_types.TaskPayload, taskCtx *TaskContext) error {
	taskCtx.LogAndUpdateStatus("Starting Docker Compose redeploy", shared_types.Cloning)

	sourceType := GetComposeSourceType(TaskPayload.Application)
	cfg := ComposeConfig{
		TaskPayload:     TaskPayload,
		TaskContext:     taskCtx,
		SourceType:      sourceType,
		ComposeFilePath: TaskPayload.Application.ComposeFilePath,
		ComposeURL:      TaskPayload.Application.ComposeFileURL,
		ComposeRaw:      TaskPayload.Application.ComposeFileContent,
	}

	composePath, err := s.PrepareComposeFile(cfg)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to prepare compose file: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.LogAndUpdateStatus("Compose file ready, rebuilding services", shared_types.Building)

	err = s.ComposeBuild(cfg, composePath)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to build compose services: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.LogAndUpdateStatus("Build complete, restarting services", shared_types.Deploying)

	err = s.ComposeRestart(cfg, composePath)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to restart compose services: "+err.Error(), shared_types.Failed)
		return err
	}

	err = s.AddComposeReverseProxy(cfg, composePath)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to configure reverse proxy: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.LogAndUpdateStatus("Docker Compose redeploy completed successfully", shared_types.Deployed)
	return nil
}

// handleDockerfileReDeploy handles redeployment of dockerfile applications
func (s *TaskService) handleDockerfileReDeploy(ctx context.Context, TaskPayload shared_types.TaskPayload, taskCtx *TaskContext) error {
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
		// Cleanup repository even on build failure
		s.CleanupRepository(repoPath)
		return err
	}

	taskCtx.AddLog("Image built successfully: " + buildImageResult + " for application " + TaskPayload.Application.Name)
	taskCtx.UpdateStatus(shared_types.Deploying)

	// Cleanup repository after build completes successfully
	s.CleanupRepository(repoPath)

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
