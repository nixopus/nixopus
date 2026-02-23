package tasks

import (
	"context"
	"strconv"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/caddy"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// HandleReDeploy clones source, builds image using redeploy flags, and atomically updates the container
func (s *TaskService) HandleReDeploy(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	taskCtx := s.NewTaskContext(TaskPayload)

	taskCtx.LogAndUpdateStatus("Starting redeploy process", shared_types.Cloning)

	repoPath, err := s.Clone(ctx, CloneConfig{
		TaskPayload:    TaskPayload,
		DeploymentType: string(shared_types.DeploymentTypeReDeploy),
		TaskContext:    taskCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to clone repository: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.LogAndUpdateStatus("Repository cloned successfully", shared_types.Building)

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
		taskCtx.LogAndUpdateStatus("Failed to build image: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.AddLog("Image built successfully: " + buildImageResult + " for application " + TaskPayload.Application.Name)

	s.ExportAndRecordImage(orgCtx, TaskPayload, buildImageResult, taskCtx)

	taskCtx.UpdateStatus(shared_types.Deploying)

	containerResult, err := s.AtomicUpdateContainer(orgCtx, TaskPayload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to update container: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.AddLog("Container updated successfully for application " + TaskPayload.Application.Name + " with container id " + containerResult.ContainerID)
	taskCtx.LogAndUpdateStatus("Redeploy completed successfully", shared_types.Deployed)

	if len(TaskPayload.Application.Domains) > 0 {
		port, err := strconv.Atoi(containerResult.AvailablePort)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to convert port to int: "+err.Error(), shared_types.Failed)
			return err
		}

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

		if err := caddy.AddDomainsAtomic(orgCtx, nil, &s.Logger, routes); err != nil {
			taskCtx.LogAndUpdateStatus("Failed to configure proxy: "+err.Error(), shared_types.Failed)
			s.cleanupServiceOnFailure(orgCtx, TaskPayload.Application.Name, taskCtx)
			return err
		}
		for _, r := range routes {
			taskCtx.AddLog("Domain " + r.Domain + " added successfully with TLS")
		}
	}

	return nil
}
