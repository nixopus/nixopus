package tasks

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type ContextTask struct {
	TaskService    *TaskService
	ContextConfig  any
	UserId         uuid.UUID
	OrganizationId uuid.UUID
	Application    *shared_types.Application
}

const (
	OperationCreate = "create"
	OperationUpdate = "update"
)

type ContextConfig struct {
	Deployment  *types.CreateDeploymentRequest
	ContextPath string
}

// GetApplicationData creates an application from a CreateDeploymentRequest
// and a user ID. It populates the application's fields with the corresponding
// values from the request, and sets the CreatedAt and UpdatedAt fields to the
// current time.
// It returns the application data.
func (c *ContextTask) GetApplicationData(
	deployment *types.CreateDeploymentRequest,
	createdAt *time.Time,
) shared_types.Application {

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
		Domain:               deployment.Domain,
		UserID:               c.UserId,
		CreatedAt:            timeValue,
		UpdatedAt:            time.Now(),
		DockerfilePath:       deployment.DockerfilePath,
		BasePath:             deployment.BasePath,
		OrganizationID:       c.OrganizationId,
	}

	return application
}

// GetDeploymentConfig creates an ApplicationDeployment from an Application.
// It sets the CreatedAt and UpdatedAt fields with the current time and returns
// the created ApplicationDeployment.
// It returns the created ApplicationDeployment.
func (c *ContextTask) GetDeploymentConfig(applicationID uuid.UUID) shared_types.ApplicationDeployment {
	applicationDeployment := shared_types.ApplicationDeployment{
		ID:              uuid.New(),
		ApplicationID:   applicationID,
		CommitHash:      "", // Initialize with an empty string
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ContainerID:     "",
		ContainerName:   "",
		ContainerImage:  "",
		ContainerStatus: "",
	}
	return applicationDeployment
}

// PersistApplicationDeploymentData persists the application and application deployment data to the database.
// It takes the operations to perform and their error messages as parameters.
// It returns an error if any operation fails.
func (c *ContextTask) PersistApplicationDeploymentData(operations []struct {
	operation  func() error
	errMessage string
}) error {
	for _, op := range operations {
		if err := c.executeDBOperations(op.operation, op.errMessage); err != nil {
			return err
		}
	}
	return nil
}

// PersistCreateApplicationDeploymentData persists the application and application deployment data to the database.
// It returns an error if the operation fails.
func (c *ContextTask) PersistCreateApplicationDeploymentData(application shared_types.Application, applicationDeployment shared_types.ApplicationDeployment) error {
	operations := []struct {
		operation  func() error
		errMessage string
	}{
		{
			operation: func() error {
				return c.TaskService.Storage.AddApplication(&application)
			},
			errMessage: types.LogFailedToCreateApplicationRecord,
		},
		{
			operation: func() error {
				return c.TaskService.Storage.AddApplicationDeployment(&applicationDeployment)
			},
			errMessage: types.LogFailedToCreateApplicationDeployment,
		},
	}
	return c.PersistApplicationDeploymentData(operations)
}

func (c *ContextTask) PersistUpdateApplicationDeploymentData(application shared_types.Application, applicationDeployment shared_types.ApplicationDeployment) error {
	operations := []struct {
		operation  func() error
		errMessage string
	}{
		{
			operation: func() error {
				return c.TaskService.Storage.UpdateApplication(&application)
			},
			errMessage: types.LogFailedToUpdateApplicationRecord,
		},
		{
			operation: func() error {
				return c.TaskService.Storage.AddApplicationDeployment(&applicationDeployment)
			},
			errMessage: types.LogFailedToUpdateApplicationDeployment,
		},
	}
	return c.PersistApplicationDeploymentData(operations)
}

// PersistApplicationDeploymentStatus creates and persists the initial application deployment status.
// It returns the created status record or an error if the operation fails.
func (c *ContextTask) PersistCreateDeploymentStatus(applicationDeployment shared_types.ApplicationDeployment) (*shared_types.ApplicationDeploymentStatus, error) {
	initialStatus := shared_types.ApplicationDeploymentStatus{
		ID:                      uuid.New(),
		ApplicationDeploymentID: applicationDeployment.ID,
		Status:                  shared_types.Started,
		UpdatedAt:               time.Now(),
	}

	err := c.TaskService.Storage.AddApplicationDeploymentStatus(&initialStatus)
	if err != nil {
		return nil, err
	}

	return &initialStatus, nil
}

// executeDBOperations executes a database operation and logs an error if it fails.
// The first parameter is a function that performs the database operation.
// The second parameter is an error message prefix that is used when logging the error.
// If the operation fails, it logs the error message and returns the error.
// Otherwise, it returns nil.
func (c *ContextTask) executeDBOperations(fn func() error, errMessage string) error {
	err := fn()
	if err != nil {
		c.TaskService.Logger.Log(logger.Error, errMessage+err.Error(), "")
		return err
	}
	return nil
}

// PrepareCreateDeploymentContext prepares the context for the deployment.
// It returns an error if the operation fails.
func (c *ContextTask) PrepareCreateDeploymentContext() (shared_types.TaskPayload, error) {
	now := time.Now()
	deployment := c.ContextConfig.(*types.CreateDeploymentRequest)
	application := c.GetApplicationData(deployment, &now)
	applicationDeployment := c.GetDeploymentConfig(application.ID)
	err := c.PersistCreateApplicationDeploymentData(application, applicationDeployment)
	if err != nil {
		return shared_types.TaskPayload{}, err
	}
	initialStatus, err := c.PersistCreateDeploymentStatus(applicationDeployment)
	if err != nil {
		return shared_types.TaskPayload{}, err
	}

	return shared_types.TaskPayload{
		Application:           application,
		ApplicationDeployment: applicationDeployment,
		Status:                initialStatus,
		UpdateOptions: shared_types.UpdateOptions{
			Force:             false,
			ForceWithoutCache: false,
		},
	}, nil
}

func (c *ContextTask) PrepareUpdateDeploymentContext() (shared_types.TaskPayload, error) {
	application := c.mergeDeploymentUpdates()
	applicationDeployment := c.GetDeploymentConfig(c.Application.ID)
	err := c.PersistUpdateApplicationDeploymentData(application, applicationDeployment)
	if err != nil {
		return shared_types.TaskPayload{}, err
	}

	initialStatus, err := c.PersistCreateDeploymentStatus(applicationDeployment)
	if err != nil {
		return shared_types.TaskPayload{}, err
	}

	return shared_types.TaskPayload{
		Application:           application,
		ApplicationDeployment: applicationDeployment,
		Status:                initialStatus,
		UpdateOptions: shared_types.UpdateOptions{
			Force:             c.ContextConfig.(*types.UpdateDeploymentRequest).Force,
			ForceWithoutCache: false, // will be set for force redeploy request for now we will not be using it
		},
	}, nil
}

// mergeDeploymentUpdates merges the updates from the deployment request into the application.
// It returns the updated application.
func (c *ContextTask) mergeDeploymentUpdates() shared_types.Application {
	deployment := c.ContextConfig.(*types.UpdateDeploymentRequest)
	application := c.Application
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
	} else {
		application.DockerfilePath = "Dockerfile"
	}

	if deployment.BasePath != "" {
		application.BasePath = deployment.BasePath
	}

	application.UpdatedAt = time.Now()

	return *application
}

func (c *ContextTask) PrepareReDeploymentContext() (shared_types.TaskPayload, error) {
	// Redeploy: keep application config, create a new deployment entry and initial status
	app := *c.Application
	app.UpdatedAt = time.Now()

	applicationDeployment := c.GetDeploymentConfig(app.ID)

	if err := c.PersistUpdateApplicationDeploymentData(app, applicationDeployment); err != nil {
		return shared_types.TaskPayload{}, err
	}

	initialStatus, err := c.PersistCreateDeploymentStatus(applicationDeployment)
	if err != nil {
		return shared_types.TaskPayload{}, err
	}

	opts := shared_types.UpdateOptions{
		Force:             c.ContextConfig.(*types.ReDeployApplicationRequest).Force,
		ForceWithoutCache: c.ContextConfig.(*types.ReDeployApplicationRequest).ForceWithoutCache,
	}

	return shared_types.TaskPayload{
		Application:           app,
		ApplicationDeployment: applicationDeployment,
		Status:                initialStatus,
		UpdateOptions:         opts,
	}, nil
}

func (c *ContextTask) PrepareRollbackContext() (shared_types.TaskPayload, error) {
	// Load the target deployment to determine commit to roll back to
	target := c.ContextConfig.(*types.RollbackDeploymentRequest)
	dep, err := c.TaskService.Storage.GetApplicationDeploymentById(target.ID.String())
	if err != nil {
		return shared_types.TaskPayload{}, err
	}

	app := *c.Application
	app.UpdatedAt = time.Now()

	applicationDeployment := c.GetDeploymentConfig(app.ID)
	applicationDeployment.CommitHash = dep.CommitHash

	if err := c.PersistUpdateApplicationDeploymentData(app, applicationDeployment); err != nil {
		return shared_types.TaskPayload{}, err
	}

	initialStatus, err := c.PersistCreateDeploymentStatus(applicationDeployment)
	if err != nil {
		return shared_types.TaskPayload{}, err
	}

	return shared_types.TaskPayload{
		Application:           app,
		ApplicationDeployment: applicationDeployment,
		Status:                initialStatus,
		UpdateOptions: shared_types.UpdateOptions{
			Force:             false,
			ForceWithoutCache: false,
		},
	}, nil
}

func (c *ContextTask) PrepareRestartContext() (shared_types.TaskPayload, error) {
	// For restart, create a fresh deployment record and initial status
	app := *c.Application
	app.UpdatedAt = time.Now()

	applicationDeployment := c.GetDeploymentConfig(app.ID)
	if err := c.PersistUpdateApplicationDeploymentData(app, applicationDeployment); err != nil {
		return shared_types.TaskPayload{}, err
	}

	initialStatus, err := c.PersistCreateDeploymentStatus(applicationDeployment)
	if err != nil {
		return shared_types.TaskPayload{}, err
	}

	return shared_types.TaskPayload{
		Application:           app,
		ApplicationDeployment: applicationDeployment,
		Status:                initialStatus,
		UpdateOptions: shared_types.UpdateOptions{
			Force:             false,
			ForceWithoutCache: false,
		},
	}, nil
}
