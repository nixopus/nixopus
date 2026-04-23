package tasks

import (
	"context"
	"strconv"

	"github.com/nixopus/nixopus/api/internal/features/deploy/caddy"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type dockerfilePipelineMode byte

const (
	dockerfilePipelineCreate dockerfilePipelineMode = iota
	dockerfilePipelineReDeploy
	dockerfilePipelineUpdate
)

func (m dockerfilePipelineMode) deploymentType() shared_types.DeploymentType {
	switch m {
	case dockerfilePipelineReDeploy:
		return shared_types.DeploymentTypeReDeploy
	case dockerfilePipelineUpdate:
		return shared_types.DeploymentTypeUpdate
	default:
		return shared_types.DeploymentTypeCreate
	}
}

func dockerfileBuildFlagsForMode(mode dockerfilePipelineMode, payload shared_types.TaskPayload) (force, forceWithoutCache bool) {
	if mode == dockerfilePipelineCreate {
		return false, false
	}
	return payload.UpdateOptions.Force, payload.UpdateOptions.ForceWithoutCache
}

func (t *TaskService) dockerfileStageResolveSource(ctx context.Context, taskCtx *TaskContext, payload shared_types.TaskPayload, mode dockerfilePipelineMode) (string, error) {
	if mode == dockerfilePipelineReDeploy {
		taskCtx.LogAndUpdateStatus("Starting redeploy process", shared_types.Cloning)
	} else {
		taskCtx.LogAndUpdateStatus("Starting deployment process", shared_types.Cloning)
	}

	resolver := t.GetSourceResolver(payload.Application.Source)
	repoPath, err := resolver.Resolve(ctx, SourceResolveConfig{
		TaskPayload:    payload,
		DeploymentType: string(mode.deploymentType()),
		TaskContext:    taskCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to resolve source: "+err.Error(), shared_types.Failed)
		return "", err
	}

	taskCtx.LogAndUpdateStatus("Source resolved successfully", shared_types.Building)
	return repoPath, nil
}

func (t *TaskService) dockerfileStageBuildImage(ctx context.Context, orgCtx context.Context, taskCtx *TaskContext, payload shared_types.TaskPayload, repoPath string, mode dockerfilePipelineMode) (string, error) {
	force, forceNoCache := dockerfileBuildFlagsForMode(mode, payload)
	taskCtx.AddLog("Building image from Dockerfile " + repoPath + " for application " + payload.Application.Name)
	buildImageResult, err := t.BuildImage(BuildConfig{
		TaskPayload:       payload,
		ContextPath:       repoPath,
		Force:             force,
		ForceWithoutCache: forceNoCache,
		TaskContext:       taskCtx,
		Context:           orgCtx,
	})
	if err != nil {
		return "", err
	}
	taskCtx.AddLog("Image built successfully: " + buildImageResult + " for application " + payload.Application.Name)
	return buildImageResult, nil
}

func (t *TaskService) dockerfileStagePublishArtifactAsync(orgCtx context.Context, taskCtx *TaskContext, payload shared_types.TaskPayload, imageTag string) {
	t.ExportAndRecordImage(orgCtx, payload, imageTag, taskCtx)
}

func (t *TaskService) dockerfileStageAtomicUpdateContainer(ctx context.Context, orgCtx context.Context, taskCtx *TaskContext, payload shared_types.TaskPayload, mode dockerfilePipelineMode) (AtomicUpdateContainerResult, error) {
	containerResult, err := t.AtomicUpdateContainer(orgCtx, payload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to update container: "+err.Error(), shared_types.Failed)
		return AtomicUpdateContainerResult{}, err
	}
	taskCtx.AddLog("Container updated successfully for application " + payload.Application.Name + " with container id " + containerResult.ContainerID)
	successMsg := "Deployment completed successfully"
	if mode == dockerfilePipelineReDeploy {
		successMsg = "Redeploy completed successfully"
	}
	taskCtx.LogAndUpdateStatus(successMsg, shared_types.Deployed)
	return containerResult, nil
}

func (t *TaskService) dockerfileStageConfigureDomains(ctx context.Context, orgCtx context.Context, taskCtx *TaskContext, payload shared_types.TaskPayload, containerResult AtomicUpdateContainerResult) error {
	if len(payload.Application.Domains) == 0 {
		return nil
	}
	port, err := strconv.Atoi(containerResult.AvailablePort)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to convert port to int: "+err.Error(), shared_types.Failed)
		return err
	}

	upstreamHost, err := GetSSHHostForOrganization(ctx, payload.Application.OrganizationID)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get SSH host: "+err.Error(), shared_types.Failed)
		return err
	}

	appDomains := make([]shared_types.ApplicationDomain, len(payload.Application.Domains))
	for i, d := range payload.Application.Domains {
		appDomains[i] = *d
	}
	routes := caddy.BuildMultiUpstreamRoutes(
		ctx, t.Storage, &t.Logger,
		payload.Application, appDomains,
		upstreamHost, port,
	)

	if err := caddy.AddDomainsAtomic(orgCtx, nil, &t.Logger, routes); err != nil {
		taskCtx.LogAndUpdateStatus("Failed to configure proxy: "+err.Error(), shared_types.Failed)
		t.cleanupServiceOnFailure(orgCtx, payload.Application.Name, taskCtx)
		return err
	}
	for _, r := range routes {
		taskCtx.AddLog("Domain " + r.Domain + " added successfully with TLS")
	}
	return nil
}

// runDockerfilePipelineFromSource runs resolve → build → async S3 export → container update → optional Caddy domains.
// Stages are ordered to match the previous HandleCreateDockerfileDeployment / redeploy / update handlers exactly.
func (t *TaskService) runDockerfilePipelineFromSource(ctx context.Context, payload shared_types.TaskPayload, mode dockerfilePipelineMode) error {
	taskCtx := t.NewTaskContext(payload)

	repoPath, err := t.dockerfileStageResolveSource(ctx, taskCtx, payload, mode)
	if err != nil {
		t.emitDeployFailed(payload, err)
		return err
	}

	if err := checkCancelled(ctx); err != nil {
		taskCtx.LogAndUpdateStatus("Deployment cancelled by user", shared_types.Cancelled)
		return err
	}

	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, payload.Application.OrganizationID.String())

	imageTag, err := t.dockerfileStageBuildImage(ctx, orgCtx, taskCtx, payload, repoPath, mode)
	if err != nil {
		if ctx.Err() != nil {
			taskCtx.LogAndUpdateStatus("Deployment cancelled by user", shared_types.Cancelled)
			return ctx.Err()
		}
		taskCtx.LogAndUpdateStatus("Failed to build image: "+err.Error(), shared_types.Failed)
		return err
	}

	if err := checkCancelled(ctx); err != nil {
		taskCtx.LogAndUpdateStatus("Deployment cancelled by user", shared_types.Cancelled)
		return err
	}

	t.dockerfileStagePublishArtifactAsync(orgCtx, taskCtx, payload, imageTag)
	taskCtx.UpdateStatus(shared_types.Deploying)

	if err := checkCancelled(ctx); err != nil {
		taskCtx.LogAndUpdateStatus("Deployment cancelled by user", shared_types.Cancelled)
		return err
	}

	containerResult, err := t.dockerfileStageAtomicUpdateContainer(ctx, orgCtx, taskCtx, payload, mode)
	if err != nil {
		t.emitDeployFailed(payload, err)
		return err
	}

	if len(payload.Application.Domains) > 0 {
		if err := t.dockerfileStageConfigureDomains(ctx, orgCtx, taskCtx, payload, containerResult); err != nil {
			t.emitDeployFailed(payload, err)
			return err
		}
	}

	return nil
}
