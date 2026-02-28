package tasks

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/caddy"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (t *TaskService) deployDockerCompose(ctx context.Context, TaskPayload shared_types.TaskPayload, deploymentType string) error {
	taskCtx := t.NewTaskContext(TaskPayload)

	repoPath, err := t.cloneRepositoryForCompose(ctx, TaskPayload, deploymentType, taskCtx)
	if err != nil {
		return err
	}

	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, TaskPayload.Application.OrganizationID.String())

	composeFilePath := t.buildComposeFilePath(TaskPayload, repoPath, taskCtx)
	envVars := GetMapFromString(TaskPayload.Application.EnvironmentVariables)
	outputCallback := t.createOutputCallback(taskCtx)

	deploymentTypeEnum := shared_types.DeploymentType(deploymentType)
	if err := t.executeComposeDeployment(orgCtx, deploymentTypeEnum, composeFilePath, envVars, outputCallback, taskCtx); err != nil {
		return err
	}

	if err := t.addDomainsForCompose(orgCtx, TaskPayload, taskCtx); err != nil {
		return err
	}

	taskCtx.LogAndUpdateStatus("Docker Compose deployment completed successfully", shared_types.Deployed)
	return nil
}

func (t *TaskService) cloneRepositoryForCompose(ctx context.Context, TaskPayload shared_types.TaskPayload, deploymentType string, taskCtx *TaskContext) (string, error) {
	taskCtx.LogAndUpdateStatus("Starting deployment process", shared_types.Cloning)

	repoPath, err := t.Clone(ctx, CloneConfig{
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

func (t *TaskService) createOutputCallback(taskCtx *TaskContext) func(string) {
	return func(line string) {
		if strings.TrimSpace(line) != "" {
			taskCtx.AddLog("Docker Compose: " + line)
		}
	}
}

func (t *TaskService) executeComposeDeployment(ctx context.Context, deploymentType shared_types.DeploymentType, composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext) error {
	switch deploymentType {
	case shared_types.DeploymentTypeCreate:
		return t.composeUp(ctx, composeFilePath, envVars, outputCallback, taskCtx, "Starting Docker Compose services", "Docker Compose services started successfully")

	case shared_types.DeploymentTypeReDeploy, shared_types.DeploymentTypeUpdate, shared_types.DeploymentTypeRollback:
		if err := t.composeDown(ctx, composeFilePath, outputCallback, taskCtx); err != nil {
			return err
		}
		taskCtx.AddLog("Existing services stopped, starting with new code")
		return t.composeUp(ctx, composeFilePath, envVars, outputCallback, taskCtx, "Starting Docker Compose services", "Docker Compose services restarted successfully")

	case shared_types.DeploymentTypeRestart:
		return t.composeRestart(ctx, composeFilePath, envVars, outputCallback, taskCtx)

	default:
		taskCtx.LogAndUpdateStatus("Unknown deployment type: "+string(deploymentType), shared_types.Failed)
		return fmt.Errorf("unknown deployment type: %s", deploymentType)
	}
}

func (t *TaskService) composeUp(ctx context.Context, composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext, startMsg, successMsg string) error {
	taskCtx.AddLog(startMsg)

	dockerSvc, err := t.getDockerService(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get docker service: "+err.Error(), shared_types.Failed)
		return err
	}

	if ds, ok := dockerSvc.(*docker.DockerService); ok {
		_, err := ds.ComposeUpWithCallback(composeFilePath, envVars, outputCallback)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to start docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	} else {
		output, err := dockerSvc.ComposeUp(composeFilePath, envVars)
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

func (t *TaskService) composeDown(ctx context.Context, composeFilePath string, outputCallback func(string), taskCtx *TaskContext) error {
	taskCtx.AddLog("Stopping existing Docker Compose services")

	dockerSvc, err := t.getDockerService(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get docker service: "+err.Error(), shared_types.Failed)
		return err
	}

	if ds, ok := dockerSvc.(*docker.DockerService); ok {
		err := ds.ComposeDownWithCallback(composeFilePath, outputCallback)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to stop docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	} else {
		err := dockerSvc.ComposeDown(composeFilePath)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to stop docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	}

	return nil
}

func (t *TaskService) composeRestart(ctx context.Context, composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext) error {
	taskCtx.AddLog("Restarting Docker Compose services")

	dockerSvc, err := t.getDockerService(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get docker service: "+err.Error(), shared_types.Failed)
		return err
	}

	if ds, ok := dockerSvc.(*docker.DockerService); ok {
		err := ds.ComposeRestart(composeFilePath, envVars, outputCallback)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to restart docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	} else {
		if err := t.composeDown(ctx, composeFilePath, outputCallback, taskCtx); err != nil {
			return err
		}
		output, err := dockerSvc.ComposeUp(composeFilePath, envVars)
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

func (t *TaskService) addDomainsForCompose(ctx context.Context, TaskPayload shared_types.TaskPayload, taskCtx *TaskContext) error {
	if len(TaskPayload.Application.Domains) == 0 {
		return nil
	}

	port := TaskPayload.Application.Port
	upstreamHost, err := GetSSHHostForOrganization(ctx, TaskPayload.Application.OrganizationID)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get SSH host: "+err.Error(), shared_types.Failed)
		return err
	}

	var routes []caddy.DomainRoute
	for _, appDomain := range TaskPayload.Application.Domains {
		if appDomain.Domain == "" {
			continue
		}
		routes = append(routes, caddy.DomainRoute{
			Domain:       appDomain.Domain,
			UpstreamDial: caddy.FormatDial(upstreamHost, port),
		})
	}

	if len(routes) == 0 {
		return nil
	}

	if err := caddy.AddDomainsAtomic(ctx, nil, &t.Logger, routes); err != nil {
		taskCtx.LogAndUpdateStatus("Failed to configure proxy: "+err.Error(), shared_types.Failed)
		return err
	}

	for _, r := range routes {
		taskCtx.AddLog("Domain " + r.Domain + " added successfully with TLS")
	}
	return nil
}
