package tasks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	apilog "github.com/nixopus/nixopus/api/internal/log"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// UpdateDeployment updates an existing application configuration
// in the database without triggering deployment
func (s *TaskService) UpdateDeployment(deployment *types.UpdateDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID) (shared_types.Application, error) {
	application, err := s.Storage.GetApplicationById(deployment.ID.String(), organizationID)
	if err != nil {
		return shared_types.Application{}, err
	}

	contextTask := ContextTask{
		TaskService:    s,
		ContextConfig:  deployment,
		UserId:         userID,
		OrganizationId: organizationID,
		Application:    &application,
	}

	// Merge the updates into the application
	updatedApplication := contextTask.mergeDeploymentUpdates()

	// Update the application in the database
	err = s.Storage.UpdateApplication(&updatedApplication)
	if err != nil {
		return shared_types.Application{}, err
	}

	// Return the updated application
	return updatedApplication, nil
}

// UpdateDeploymentWithTrigger updates an existing application configuration
// in the database and triggers the deployment process
// This is used for webhooks and other cases where deployment should be triggered
func (s *TaskService) UpdateDeploymentWithTrigger(deployment *types.UpdateDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID) (shared_types.Application, error) {
	application, err := s.Storage.GetApplicationById(deployment.ID.String(), organizationID)
	if err != nil {
		return shared_types.Application{}, err
	}

	contextTask := ContextTask{
		TaskService:    s,
		ContextConfig:  deployment,
		UserId:         userID,
		OrganizationId: organizationID,
		Application:    &application,
	}

	TaskPayload, err := contextTask.PrepareUpdateDeploymentContext()
	if err != nil {
		return shared_types.Application{}, err
	}

	TaskPayload.CorrelationID = uuid.NewString()

	err = UpdateDeploymentQueue.Add(TaskUpdateDeployment.WithArgs(context.Background(), TaskPayload))
	if err != nil {
		apilog.Errorf("error enqueuing update deployment: %v", err)
		return shared_types.Application{}, err
	}

	return application, nil
}

// HandleUpdateDeployment routes update deployment based on the application's BuildPack type
func (s *TaskService) HandleUpdateDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	switch TaskPayload.Application.BuildPack {
	case shared_types.DockerFile:
		return s.HandleUpdateDockerfileDeployment(ctx, TaskPayload)
	case shared_types.DockerCompose:
		return s.HandleUpdateDockerComposeDeployment(ctx, TaskPayload)
	case shared_types.Static:
		return s.HandleUpdateStaticDeployment(ctx, TaskPayload)
	default:
		return types.ErrInvalidBuildPack
	}
}

// HandleUpdateDockerfileDeployment handles update deployment of a Dockerfile-based application
func (s *TaskService) HandleUpdateDockerfileDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return s.runDockerfilePipelineFromSource(ctx, TaskPayload, dockerfilePipelineUpdate)
}

// HandleUpdateDockerComposeDeployment handles update deployment of a Docker Compose application
func (s *TaskService) HandleUpdateDockerComposeDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	return s.deployDockerCompose(ctx, TaskPayload, string(shared_types.DeploymentTypeUpdate))
}

// HandleUpdateStaticDeployment handles update deployment of a static application
func (s *TaskService) HandleUpdateStaticDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	// TODO: Implement static update deployment
	return fmt.Errorf("static update deployment not yet implemented")
}
