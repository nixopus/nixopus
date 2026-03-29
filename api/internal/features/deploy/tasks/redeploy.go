package tasks

import (
	"context"
	"strconv"

	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/deploy/caddy"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// HandleReDeploy fans out redeployment across all configured servers (or org default for single-server apps).
func (s *TaskService) HandleReDeploy(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	allServers, err := s.Storage.GetApplicationServers(TaskPayload.Application.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve application servers: %w", err)
	}
	if len(allServers) == 0 {
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
	taskCtx := s.NewTaskContext(TaskPayload)

	taskCtx.LogAndUpdateStatus("Starting redeploy process", shared_types.Cloning)

	resolver := s.GetSourceResolver(TaskPayload.Application.Source)
	repoPath, err := resolver.Resolve(ctx, SourceResolveConfig{
		TaskPayload:    TaskPayload,
		DeploymentType: string(shared_types.DeploymentTypeReDeploy),
		TaskContext:    taskCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to resolve source: "+err.Error(), shared_types.Failed)
		s.emitDeployFailed(TaskPayload, err)
		return err
	}

	taskCtx.LogAndUpdateStatus("Source resolved successfully", shared_types.Building)

	if err := checkCancelled(ctx); err != nil {
		taskCtx.LogAndUpdateStatus("Deployment cancelled by user", shared_types.Cancelled)
		return err
	}

	// Add organization ID to context for docker service
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, TaskPayload.Application.OrganizationID.String())

	taskCtx.AddLog("Building image from Dockerfile " + repoPath + " for application " + TaskPayload.Application.Name)

	buildImageResult, err := s.BuildImage(BuildConfig{
		TaskPayload:       TaskPayload,
		ContextPath:       repoPath,
		Force:             TaskPayload.UpdateOptions.Force,
		ForceWithoutCache: TaskPayload.UpdateOptions.ForceWithoutCache,
		TaskContext:       taskCtx,
		Context:           orgCtx,
	})
	if err != nil {
		if ctx.Err() != nil {
			taskCtx.LogAndUpdateStatus("Deployment cancelled by user", shared_types.Cancelled)
			return ctx.Err()
		}
		taskCtx.LogAndUpdateStatus("Failed to build image: "+err.Error(), shared_types.Failed)
		s.emitDeployFailed(TaskPayload, err)
		return err
	}

	if err := checkCancelled(ctx); err != nil {
		taskCtx.LogAndUpdateStatus("Deployment cancelled by user", shared_types.Cancelled)
		return err
	}

	taskCtx.AddLog("Image built successfully: " + buildImageResult + " for application " + TaskPayload.Application.Name)

	s.ExportAndRecordImage(orgCtx, TaskPayload, buildImageResult, taskCtx)

	taskCtx.UpdateStatus(shared_types.Deploying)

	if err := checkCancelled(ctx); err != nil {
		taskCtx.LogAndUpdateStatus("Deployment cancelled by user", shared_types.Cancelled)
		return err
	}

	containerResult, err := s.AtomicUpdateContainer(orgCtx, TaskPayload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to update container: "+err.Error(), shared_types.Failed)
		s.emitDeployFailed(TaskPayload, err)
		return err
	}

	taskCtx.AddLog("Container updated successfully for application " + TaskPayload.Application.Name + " with container id " + containerResult.ContainerID)
	taskCtx.LogAndUpdateStatus("Redeploy completed successfully", shared_types.Deployed)

	if len(TaskPayload.Application.Domains) > 0 {
		port, err := strconv.Atoi(containerResult.AvailablePort)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to convert port to int: "+err.Error(), shared_types.Failed)
			s.emitDeployFailed(TaskPayload, err)
			return err
		}

		upstreamHost, err := GetSSHHostForOrganization(ctx, TaskPayload.Application.OrganizationID)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to get SSH host: "+err.Error(), shared_types.Failed)
			s.emitDeployFailed(TaskPayload, err)
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

		if err := caddy.AddDomainsAtomic(orgCtx, nil, &s.Logger, routes); err != nil {
			taskCtx.LogAndUpdateStatus("Failed to configure proxy: "+err.Error(), shared_types.Failed)
			s.emitDeployFailed(TaskPayload, err)
			s.cleanupServiceOnFailure(orgCtx, TaskPayload.Application.Name, taskCtx)
			return err
		}
		for _, r := range routes {
			taskCtx.AddLog("Domain " + r.Domain + " added successfully with TLS")
		}
	}

	return nil
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
