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
		fmt.Printf("error enqueuing update deployment: %v\n", err)
		return shared_types.Application{}, err
	}

	return application, nil
}

func (s *TaskService) HandleUpdateDeployment(ctx context.Context, TaskPayload shared_types.TaskPayload) error {
	taskCtx := s.NewTaskContext(TaskPayload)

	taskCtx.LogAndUpdateStatus("Starting deployment process", shared_types.Cloning)

	repoPath, err := s.Clone(CloneConfig{
		TaskPayload:    TaskPayload,
		DeploymentType: string(shared_types.DeploymentTypeUpdate),
		TaskContext:    taskCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to clone repository: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.LogAndUpdateStatus("Repository cloned successfully", shared_types.Building)
	taskCtx.AddLog("Building image from Dockerfile " + repoPath + " for application " + TaskPayload.Application.Name)
	buildImageResult, err := s.BuildImage(BuildConfig{
		TaskPayload:       TaskPayload,
		ContextPath:       repoPath,
		Force:             TaskPayload.UpdateOptions.Force,
		ForceWithoutCache: TaskPayload.UpdateOptions.ForceWithoutCache,
		TaskContext:       taskCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to build image: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.AddLog("Image built successfully: " + buildImageResult + " for application " + TaskPayload.Application.Name)
	taskCtx.UpdateStatus(shared_types.Deploying)

	containerResult, err := s.AtomicUpdateContainer(TaskPayload, taskCtx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to update container: "+err.Error(), shared_types.Failed)
		return err
	}

	taskCtx.AddLog("Container updated successfully for application " + TaskPayload.Application.Name + " with container id " + containerResult.ContainerID)
	taskCtx.LogAndUpdateStatus("Deployment completed successfully", shared_types.Deployed)

	// Add domains to proxy if any are provided
	if len(TaskPayload.Application.Domains) > 0 {
		client := GetCaddyClient()
		port, err := strconv.Atoi(containerResult.AvailablePort)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to convert port to int: "+err.Error(), shared_types.Failed)
			return err
		}
		upstreamHost := config.AppConfig.SSH.Host

		// Loop through all domains and add them with TLS
		for _, appDomain := range TaskPayload.Application.Domains {
			if appDomain.Domain == "" {
				continue
			}
			err = client.AddDomainWithAutoTLS(appDomain.Domain, upstreamHost, port, caddygo.DomainOptions{})
			if err != nil {
				fmt.Printf("Failed to add domain %s: %v\n", appDomain.Domain, err)
				taskCtx.LogAndUpdateStatus("Failed to add domain "+appDomain.Domain+": "+err.Error(), shared_types.Failed)
				return err
			}
			taskCtx.AddLog("Domain " + appDomain.Domain + " added successfully with TLS")
		}
		client.Reload()
	}

	return nil
}
