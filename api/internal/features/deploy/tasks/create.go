package tasks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
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

func (t *TaskService) CreateTemplateDeploymentTask(deployment *types.CreateDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID, templateID string) (shared_types.Application, error) {
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

	TaskPayload.Application.TemplateID = templateID
	if _, err := t.Store.DB.NewUpdate().
		Model(&TaskPayload.Application).
		Column("template_id").
		Where("id = ?", TaskPayload.Application.ID).
		Exec(context.Background()); err != nil {
		return shared_types.Application{}, fmt.Errorf("failed to set template_id: %w", err)
	}

	TaskPayload.CorrelationID = uuid.NewString()

	err = CreateDeploymentQueue.Add(TaskCreateDeployment.WithArgs(context.Background(), TaskPayload))
	if err != nil {
		return shared_types.Application{}, fmt.Errorf("failed to enqueue deployment: %w", err)
	}

	return TaskPayload.Application, nil
}

func (t *TaskService) HandleCreateDockerfileDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return t.runDockerfilePipelineFromSource(ctx, TaskPayload, dockerfilePipelineCreate)
}

// HandleCreateDockerComposeDeployment handles the deployment of a Docker Compose application
func (t *TaskService) HandleCreateDockerComposeDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return t.deployDockerCompose(ctx, TaskPayload, string(shared_types.DeploymentTypeCreate))
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
