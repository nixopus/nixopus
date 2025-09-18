package tasks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
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
		fmt.Printf("error enqueuing create deployment: %v\n", err)
	}

	return TaskPayload.Application, nil
}

func (t *TaskService) HandleCreateDockerfileDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	taskCtx := t.NewTaskContext(TaskPayload)

	taskCtx.LogAndUpdateStatus("Starting deployment process", shared_types.Cloning)

	repoPath, err := t.Clone(CloneConfig{
		TaskPayload:    TaskPayload,
		DeploymentType: string(shared_types.DeploymentTypeCreate),
		TaskContext:    taskCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to clone repository: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.LogAndUpdateStatus("Repository cloned successfully", shared_types.Building)
	taskCtx.AddLog("Building image from Dockerfile " + repoPath + " for application " + TaskPayload.Application.Name)
	buildImageResult, err := t.BuildImage(BuildConfig{
		TaskPayload:       TaskPayload,
		ContextPath:       repoPath,
		Force:             false,
		ForceWithoutCache: false,
		TaskContext:       taskCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to build image: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.AddLog("Image built successfully: " + buildImageResult + " for application " + TaskPayload.Application.Name)
	taskCtx.UpdateStatus(shared_types.Deploying)

	containerResult, err := t.AtomicUpdateContainer(TaskPayload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to update container: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.AddLog("Container updated successfully for application " + TaskPayload.Application.Name + " with container id " + containerResult.ContainerID)
	taskCtx.LogAndUpdateStatus("Deployment completed successfully", shared_types.Deployed)

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

// TODO : Implement the docker compose deployment
func (t *TaskService) HandleCreateDockerComposeDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return nil
}

// TODO : Implement the static deployment
func (t *TaskService) HandleCreateStaticDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return nil
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
		fmt.Printf("error enqueuing redeploy: %v\n", err)
	}

	return application, nil
}
