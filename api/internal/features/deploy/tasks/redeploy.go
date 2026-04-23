package tasks

import (
	"context"
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// HandleReDeploy fans out redeployment across all configured servers (or org default for single-server apps).
func (s *TaskService) HandleReDeploy(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	s.Logger.Log(logger.Info, fmt.Sprintf("redeploy: fetching servers for app %s (buildpack=%s)", TaskPayload.Application.ID, TaskPayload.Application.BuildPack), "")
	allServers, err := s.Storage.GetApplicationServers(TaskPayload.Application.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve application servers: %w", err)
	}
	s.Logger.Log(logger.Info, fmt.Sprintf("redeploy: got %d servers for app %s", len(allServers), TaskPayload.Application.ID), "")
	if len(allServers) == 0 {
		s.Logger.Log(logger.Info, "redeploy: no application_servers rows, falling back to org default", "")
		return s.handleReDeploySingle(ctx, TaskPayload)
	}
	servers := filterServers(allServers, TaskPayload.TargetServerIDs)
	if len(servers) == 0 && len(TaskPayload.TargetServerIDs) > 0 {
		return fmt.Errorf("none of the requested target servers are assigned to this application")
	}
	if len(servers) == 0 {
		servers = allServers
	}
	if len(servers) == 1 {
		return s.handleReDeploySingle(ctx, TaskPayload)
	}
	return s.fanOut(ctx, TaskPayload, servers, s.handleReDeploySingle)
}

// handleReDeploySingle routes redeployment based on the application's BuildPack type.
func (s *TaskService) handleReDeploySingle(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	switch TaskPayload.Application.BuildPack {
	case shared_types.DockerFile:
		return s.HandleReDeployDockerfileDeployment(ctx, TaskPayload)
	case shared_types.DockerCompose:
		return s.HandleReDeployDockerComposeDeployment(ctx, TaskPayload)
	case shared_types.Static:
		return s.HandleReDeployStaticDeployment(ctx, TaskPayload)
	default:
		return types.ErrInvalidBuildPack
	}
}

// HandleReDeployDockerfileDeployment handles redeployment of a Dockerfile-based application
func (s *TaskService) HandleReDeployDockerfileDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return s.runDockerfilePipelineFromSource(ctx, TaskPayload, dockerfilePipelineReDeploy)
}

// HandleReDeployDockerComposeDeployment handles redeployment of a Docker Compose application
func (s *TaskService) HandleReDeployDockerComposeDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return s.deployDockerCompose(ctx, TaskPayload, string(shared_types.DeploymentTypeReDeploy))
}

// HandleReDeployStaticDeployment handles redeployment of a static application
func (s *TaskService) HandleReDeployStaticDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	// TODO: Implement static redeployment
	return fmt.Errorf("static redeployment not yet implemented")
}
