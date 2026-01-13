package tasks

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// deployDockerCompose handles the common logic for docker compose deployment
func (t *TaskService) deployDockerCompose(ctx context.Context, TaskPayload shared_types.TaskPayload, deploymentType string) error {
	taskCtx := t.NewTaskContext(TaskPayload)

	// Clone repository
	repoPath, err := t.cloneRepositoryForCompose(TaskPayload, deploymentType, taskCtx)
	if err != nil {
		return err
	}

	// Build compose file path
	composeFilePath := t.buildComposeFilePath(TaskPayload, repoPath, taskCtx)
	envVars := GetMapFromString(TaskPayload.Application.EnvironmentVariables)
	outputCallback := t.createOutputCallback(taskCtx)

	// Execute deployment based on type
	deploymentTypeEnum := shared_types.DeploymentType(deploymentType)
	if err := t.executeComposeDeployment(deploymentTypeEnum, composeFilePath, envVars, outputCallback, taskCtx); err != nil {
		return err
	}

	// Add domain if specified
	if err := t.addDomainForCompose(TaskPayload, taskCtx); err != nil {
		return err
	}

	taskCtx.LogAndUpdateStatus("Docker Compose deployment completed successfully", shared_types.Deployed)
	return nil
}

// cloneRepositoryForCompose clones the repository for compose deployment
func (t *TaskService) cloneRepositoryForCompose(TaskPayload shared_types.TaskPayload, deploymentType string, taskCtx *TaskContext) (string, error) {
	taskCtx.LogAndUpdateStatus("Starting deployment process", shared_types.Cloning)

	repoPath, err := t.Clone(CloneConfig{
		TaskPayload:    TaskPayload,
		DeploymentType: deploymentType,
		TaskContext:    taskCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to clone repository: "+err.Error(), shared_types.Failed)
		return "", err
	}

	taskCtx.LogAndUpdateStatus("Repository cloned successfully", shared_types.Deploying)
	taskCtx.AddLog("Repository cloned to: " + repoPath)
	return repoPath, nil
}

// buildComposeFilePath builds the path to the docker-compose file
func (t *TaskService) buildComposeFilePath(TaskPayload shared_types.TaskPayload, repoPath string, taskCtx *TaskContext) string {
	basePath := TaskPayload.Application.BasePath
	if basePath == "" || basePath == "/" {
		basePath = "."
	}

	composeFileName := "docker-compose.yml"
	if TaskPayload.Application.DockerfilePath != "" && TaskPayload.Application.DockerfilePath != "Dockerfile" {
		composeFileName = TaskPayload.Application.DockerfilePath
	}

	composeFilePath := filepath.Join(repoPath, basePath, composeFileName)
	taskCtx.AddLog("Starting Docker Compose services from: " + composeFilePath)
	return composeFilePath
}

// createOutputCallback creates a callback function for streaming output
func (t *TaskService) createOutputCallback(taskCtx *TaskContext) func(string) {
	return func(line string) {
		if strings.TrimSpace(line) != "" {
			taskCtx.AddLog("Docker Compose: " + line)
		}
	}
}

// executeComposeDeployment executes the appropriate compose command based on deployment type
func (t *TaskService) executeComposeDeployment(deploymentType shared_types.DeploymentType, composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext) error {
	switch deploymentType {
	case shared_types.DeploymentTypeCreate:
		return t.composeUp(composeFilePath, envVars, outputCallback, taskCtx, "Starting Docker Compose services", "Docker Compose services started successfully")

	case shared_types.DeploymentTypeReDeploy, shared_types.DeploymentTypeUpdate, shared_types.DeploymentTypeRollback:
		if err := t.composeDown(composeFilePath, outputCallback, taskCtx); err != nil {
			return err
		}
		taskCtx.AddLog("Existing services stopped, starting with new code")
		return t.composeUp(composeFilePath, envVars, outputCallback, taskCtx, "Starting Docker Compose services", "Docker Compose services restarted successfully")

	case shared_types.DeploymentTypeRestart:
		return t.composeRestart(composeFilePath, envVars, outputCallback, taskCtx)

	default:
		taskCtx.LogAndUpdateStatus("Unknown deployment type: "+string(deploymentType), shared_types.Failed)
		return fmt.Errorf("unknown deployment type: %s", deploymentType)
	}
}

// composeUp starts docker compose services
func (t *TaskService) composeUp(composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext, startMsg, successMsg string) error {
	taskCtx.AddLog(startMsg)

	if dockerSvc, ok := t.DockerRepo.(*docker.DockerService); ok {
		_, err := dockerSvc.ComposeUpWithCallback(composeFilePath, envVars, outputCallback)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to start docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	} else {
		output, err := t.DockerRepo.ComposeUp(composeFilePath, envVars)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to start docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
		if output != "" {
			taskCtx.AddLog("Docker Compose output: " + output)
		}
	}

	taskCtx.AddLog(successMsg)
	return nil
}

// composeDown stops docker compose services
func (t *TaskService) composeDown(composeFilePath string, outputCallback func(string), taskCtx *TaskContext) error {
	taskCtx.AddLog("Stopping existing Docker Compose services")

	if dockerSvc, ok := t.DockerRepo.(*docker.DockerService); ok {
		err := dockerSvc.ComposeDownWithCallback(composeFilePath, outputCallback)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to stop docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	} else {
		err := t.DockerRepo.ComposeDown(composeFilePath)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to stop docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	}

	return nil
}

// composeRestart restarts docker compose services
func (t *TaskService) composeRestart(composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext) error {
	taskCtx.AddLog("Restarting Docker Compose services")

	if dockerSvc, ok := t.DockerRepo.(*docker.DockerService); ok {
		err := dockerSvc.ComposeRestart(composeFilePath, envVars, outputCallback)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to restart docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	} else {
		// Fallback: use down + up for restart
		if err := t.composeDown(composeFilePath, outputCallback, taskCtx); err != nil {
			return err
		}
		output, err := t.DockerRepo.ComposeUp(composeFilePath, envVars)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to start docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
		if output != "" {
			taskCtx.AddLog("Docker Compose output: " + output)
		}
	}

	taskCtx.AddLog("Docker Compose services restarted successfully")
	return nil
}

// addDomainForCompose adds domain configuration if specified
func (t *TaskService) addDomainForCompose(TaskPayload shared_types.TaskPayload, taskCtx *TaskContext) error {
	if TaskPayload.Application.Domain == "" {
		return nil
	}

	client := GetCaddyClient()
	port := TaskPayload.Application.Port
	upstreamHost := config.AppConfig.SSH.Host

	taskCtx.AddLog(fmt.Sprintf("Adding domain %s pointing to %s:%d", TaskPayload.Application.Domain, upstreamHost, port))

	err := client.AddDomainWithAutoTLS(TaskPayload.Application.Domain, upstreamHost, port, caddygo.DomainOptions{})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to add domain: "+err.Error(), shared_types.Failed)
		return err
	}

	client.Reload()
	taskCtx.AddLog("Domain added successfully: " + TaskPayload.Application.Domain)
	return nil
}
