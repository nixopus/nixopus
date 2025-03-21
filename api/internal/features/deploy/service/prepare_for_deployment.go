package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// createApplicationFromDeploymentRequest creates an application from a CreateDeploymentRequest
// and a user ID. It populates the application's fields with the corresponding
// values from the request, and sets the CreatedAt and UpdatedAt fields to the
// current time.
func createApplicationFromDeploymentRequest(deployment *types.CreateDeploymentRequest, userID uuid.UUID, createdAt *time.Time) shared_types.Application {
	timeValue := time.Now()
	if createdAt != nil {
		timeValue = *createdAt
	}

	application := shared_types.Application{
		ID:                   uuid.New(),
		Name:                 deployment.Name,
		BuildVariables:       GetStringFromMap(deployment.BuildVariables),
		EnvironmentVariables: GetStringFromMap(deployment.EnvironmentVariables),
		Environment:          deployment.Environment,
		BuildPack:            deployment.BuildPack,
		Repository:           deployment.Repository,
		Branch:               deployment.Branch,
		PreRunCommand:        deployment.PreRunCommand,
		PostRunCommand:       deployment.PostRunCommand,
		Port:                 deployment.Port,
		DomainID:             deployment.DomainID,
		UserID:               userID,
		CreatedAt:            timeValue,
		UpdatedAt:            time.Now(),
		DockerfilePath:       deployment.DockerfilePath,
	}

	return application
}

func createApplicationFromExistingApplicationAndUpdateRequest(application shared_types.Application, deployment *types.UpdateDeploymentRequest) shared_types.Application {
	if deployment.Name != "" {
		application.Name = deployment.Name
	}

	if deployment.BuildVariables != nil {
		application.BuildVariables = GetStringFromMap(deployment.BuildVariables)
	}

	if deployment.EnvironmentVariables != nil {
		application.EnvironmentVariables = GetStringFromMap(deployment.EnvironmentVariables)
	}

	if deployment.PreRunCommand != "" {
		application.PreRunCommand = deployment.PreRunCommand
	}

	if deployment.PostRunCommand != "" {
		application.PostRunCommand = deployment.PostRunCommand
	}

	if deployment.Port != 0 {
		application.Port = deployment.Port
	}

	if deployment.DockerfilePath != "" {
		application.DockerfilePath = deployment.DockerfilePath
	}else{
		application.DockerfilePath = "Dockerfile"
	}

	application.UpdatedAt = time.Now()

	return application
}

// createDeploymentConfig creates an ApplicationDeployment from an Application.
// It sets the CreatedAt and UpdatedAt fields with the current time and returns
// the created ApplicationDeployment.
func createDeploymentConfig(application shared_types.Application) shared_types.ApplicationDeployment {
	deployment_config := shared_types.ApplicationDeployment{
		ID:              uuid.New(),
		ApplicationID:   application.ID,
		CommitHash:      "", // Initialize with an empty string
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ContainerID:     "",
		ContainerName:   "",
		ContainerImage:  "",
		ContainerStatus: "",
	}

	return deployment_config
}

// createAppLogs creates an ApplicationLogs with the given application ID and the log
// message LogDeploymentStarted. The CreatedAt and UpdatedAt fields are set
// to the current time.
func createAppLogs(application shared_types.Application, ApplicationDeploymentID uuid.UUID) shared_types.ApplicationLogs {
	app_logs := shared_types.ApplicationLogs{
		ID:                      uuid.New(),
		ApplicationID:           application.ID,
		Log:                     types.LogDeploymentStarted,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		ApplicationDeploymentID: ApplicationDeploymentID,
	}

	return app_logs
}

// createDeploymentStatus creates an ApplicationDeploymentStatus from an ApplicationDeployment.
// It sets the Status to Started and populates the CreatedAt and UpdatedAt fields
// with the current time.
func createDeploymentStatus(ApplicationDeploymentID uuid.UUID) shared_types.ApplicationDeploymentStatus {
	deployment_status := shared_types.ApplicationDeploymentStatus{
		ID:                      uuid.New(),
		ApplicationDeploymentID: ApplicationDeploymentID,
		Status:                  shared_types.Started,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	return deployment_status
}

// createAndPrepareDeployment handles the common logic for creating, updating, and redeploying applications
// It creates or updates application records, deployment configs, logs, and statuses
func (s *DeployService) createAndPrepareDeployment(application shared_types.Application, d shared_types.DeploymentRequestConfig) (*shared_types.DeploymentRequestConfig, shared_types.ApplicationDeploymentStatus, shared_types.ApplicationDeployment, error) {
	deployment_config := createDeploymentConfig(application)
	appLogs := createAppLogs(application, deployment_config.ID)
	deployment_status := createDeploymentStatus(deployment_config.ID)

	operations := []struct {
		operation  func() error
		errMessage string
	}{
		{
			operation: func() error {
				// if the deployment type is create, add the application to the database else we will update the application in case of redeploy or update or any other type
				if d.Type == shared_types.DeploymentTypeCreate {
					return s.storage.AddApplication(&application)
				}
				return s.storage.UpdateApplication(&application)
			},
			errMessage: types.LogFailedToCreateApplicationRecord,
		},
		{
			operation: func() error {
				return s.storage.AddApplicationDeployment(&deployment_config)
			},
			errMessage: types.LogFailedToCreateApplicationDeployment,
		},
		{
			operation: func() error {
				return s.storage.AddApplicationLogs(&appLogs)
			},
			errMessage: types.LogFailedToCreateApplicationLogs,
		},
		{
			operation: func() error {
				return s.storage.AddApplicationDeploymentStatus(&deployment_status)
			},
			errMessage: types.LogFailedToCreateApplicationDeploymentStatus,
		},
	}

	for _, op := range operations {
		if err := s.executeDBOperations(op.operation, op.errMessage); err != nil {
			return &shared_types.DeploymentRequestConfig{}, shared_types.ApplicationDeploymentStatus{}, shared_types.ApplicationDeployment{}, err
		}
	}

	deploymentRequest := shared_types.DeploymentRequestConfig{
		Type:              d.Type,
		Force:             d.Force,
		ForceWithoutCache: d.ForceWithoutCache,
	}

	return &deploymentRequest, deployment_status, deployment_config, nil
}

// executeDBOperations executes a database operation and logs an error if it fails.
// The first parameter is a function that performs the database operation.
// The second parameter is an error message prefix that is used when logging the error.
// If the operation fails, it logs the error message and returns the error.
// Otherwise, it returns nil.
func (s *DeployService) executeDBOperations(fn func() error, errMessage string) error {
	err := fn()
	if err != nil {
		s.logger.Log(logger.Error, errMessage+err.Error(), "")
		return err
	}
	return nil
}
