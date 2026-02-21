package tasks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/caddy"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (t *TaskService) CreateDeploymentTask(deployment *types.CreateDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID) (shared_types.Application, error) {
	contextTask := ContextTask{
		TaskService:    t,
		ContextConfig:  deployment,
		UserId:         userID,
		OrganizationId: organizationID,
	}

	TaskPayload, err := contextTask.PrepareCreateDeploymentContext()
	if err != nil {
		return shared_types.Application{}, err
	}

	TaskPayload.CorrelationID = uuid.NewString()

	err = CreateDeploymentQueue.Add(TaskCreateDeployment.WithArgs(context.Background(), TaskPayload))
	if err != nil {
		return shared_types.Application{}, fmt.Errorf("failed to enqueue deployment: %w", err)
	}

	return TaskPayload.Application, nil
}

func (t *TaskService) HandleCreateDockerfileDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	taskCtx := t.NewTaskContext(TaskPayload)

	taskCtx.LogAndUpdateStatus("Starting deployment process", shared_types.Cloning)

	repoPath, err := t.Clone(ctx, CloneConfig{
		TaskPayload:    TaskPayload,
		DeploymentType: string(shared_types.DeploymentTypeCreate),
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
	buildImageResult, err := t.BuildImage(BuildConfig{
		TaskPayload:       TaskPayload,
		ContextPath:       repoPath,
		Force:             false,
		ForceWithoutCache: false,
		TaskContext:       taskCtx,
		Context:           orgCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to build image: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.AddLog("Image built successfully: " + buildImageResult + " for application " + TaskPayload.Application.Name)

	t.ExportAndRecordImage(orgCtx, TaskPayload, buildImageResult, taskCtx)

	taskCtx.UpdateStatus(shared_types.Deploying)

	containerResult, err := t.AtomicUpdateContainer(orgCtx, TaskPayload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to update container: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.AddLog("Container updated successfully for application " + TaskPayload.Application.Name + " with container id " + containerResult.ContainerID)
	taskCtx.LogAndUpdateStatus("Deployment completed successfully", shared_types.Deployed)

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

		if err := caddy.AddDomainsAtomic(orgCtx, nil, &t.Logger, routes); err != nil {
			taskCtx.LogAndUpdateStatus("Failed to configure proxy: "+err.Error(), shared_types.Failed)
			return err
		}
		for _, r := range routes {
			taskCtx.AddLog("Domain " + r.Domain + " added successfully with TLS")
		}
	}
	return nil
}

// TODO : Implement the docker compose deployment
func (t *TaskService) HandleCreateDockerComposeDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return nil
}

// TODO : Implement the static deployment
func (t *TaskService) HandleCreateStaticDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return nil
}

// DeployProject triggers deployment of an existing project (application) that was saved as a draft.
func (t *TaskService) DeployProject(request *types.DeployProjectRequest, userID uuid.UUID, organizationID uuid.UUID) (shared_types.Application, error) {
	application, err := t.Storage.GetApplicationById(request.ID.String(), organizationID)
	if err != nil {
		return shared_types.Application{}, types.ErrApplicationNotFound
	}

	// Check if the application is in draft status (no deployments yet)
	if application.Status != nil && application.Status.Status != shared_types.Draft {
		return shared_types.Application{}, types.ErrApplicationNotDraft
	}

	contextTask := ContextTask{
		TaskService:    t,
		ContextConfig:  request,
		UserId:         userID,
		OrganizationId: organizationID,
		Application:    &application,
	}

	TaskPayload, err := contextTask.PrepareDeployProjectContext()
	if err != nil {
		return shared_types.Application{}, err
	}

	TaskPayload.CorrelationID = uuid.NewString()

	err = CreateDeploymentQueue.Add(TaskCreateDeployment.WithArgs(context.Background(), TaskPayload))
	if err != nil {
		return shared_types.Application{}, fmt.Errorf("failed to enqueue project deployment: %w", err)
	}

	return application, nil
}

// TODOD: Shravan implement types and get back
func (t *TaskService) ReDeployApplication(request *types.ReDeployApplicationRequest, userID uuid.UUID, organizationID uuid.UUID) (shared_types.Application, error) {
	application, err := t.Storage.GetApplicationById(request.ID.String(), organizationID)
	if err != nil {
		return shared_types.Application{}, err
	}

	contextTask := ContextTask{
		TaskService:    t,
		ContextConfig:  request,
		UserId:         userID,
		OrganizationId: organizationID,
		Application:    &application,
	}

	TaskPayload, err := contextTask.PrepareReDeploymentContext()
	if err != nil {
		return shared_types.Application{}, err
	}

	TaskPayload.CorrelationID = uuid.NewString()

	err = ReDeployQueue.Add(TaskReDeploy.WithArgs(context.Background(), TaskPayload))
	if err != nil {
		return shared_types.Application{}, fmt.Errorf("failed to enqueue redeploy: %w", err)
	}

	return application, nil
}
