package types

import (
	"errors"

	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type IsNameAlreadyTakenRequest struct {
	Name string `json:"name"`
}

type IsDomainAlreadyTakenRequest struct {
	Domain string `json:"domain"`
}

type IsDomainValidRequest struct {
	Domain string `json:"domain"`
}

type IsPortAlreadyTakenRequest struct {
	Port int `json:"port"`
}

type CreateDeploymentRequest struct {
	Name                 string                   `json:"name"`
	Domain               string                   `json:"domain"`
	Environment          shared_types.Environment `json:"environment"`
	BuildPack            shared_types.BuildPack   `json:"build_pack"`
	Repository           string                   `json:"repository"`
	Branch               string                   `json:"branch"`
	PreRunCommand        string                   `json:"pre_run_command"`
	PostRunCommand       string                   `json:"post_run_command"`
	BuildVariables       map[string]string        `json:"build_variables"`
	EnvironmentVariables map[string]string        `json:"environment_variables"`
	Port                 int                      `json:"port"`
	DockerfilePath       string                   `json:"dockerfile_path,omitempty"`
	BasePath             string                   `json:"base_path,omitempty"`
}

type UpdateDeploymentRequest struct {
	Name                 string            `json:"name,omitempty"`
	PreRunCommand        string            `json:"pre_run_command,omitempty"`
	PostRunCommand       string            `json:"post_run_command,omitempty"`
	BuildVariables       map[string]string `json:"build_variables,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
	Port                 int               `json:"port,omitempty"`
	ID                   uuid.UUID         `json:"id,omitempty"`
	Force                bool              `json:"force,omitempty"`
	DockerfilePath       string            `json:"dockerfile_path,omitempty"`
	BasePath             string            `json:"base_path,omitempty"`
}

type DeleteDeploymentRequest struct {
	ID uuid.UUID `json:"id"`
}

type ReDeployApplicationRequest struct {
	ID                uuid.UUID `json:"id"`
	Force             bool      `json:"force"`
	ForceWithoutCache bool      `json:"force_without_cache"`
}

type RollbackDeploymentRequest struct {
	ID uuid.UUID `json:"id"`
}

type RestartDeploymentRequest struct {
	ID uuid.UUID `json:"id"`
}

var (
	ErrMissingID                    = errors.New("id is required")
	ErrInvalidRequestType           = errors.New("invalid request type")
	ErrMissingName                  = errors.New("name is required")
	ErrMissingDomain                = errors.New("domain is required")
	ErrMissingRepository            = errors.New("repository is required")
	ErrMissingBranch                = errors.New("branch is required")
	ErrMissingPort                  = errors.New("port is required")
	ErrInvalidEnvironment           = errors.New("invalid environment")
	ErrInvalidBuildPack             = errors.New("invalid build pack")
	ErrFailedToCreateTarFromContext = errors.New("failed to create tar from context")
	ErrProcessingBuildOutput        = errors.New("failed to process build output")
	ErrBuildDockerImage             = errors.New("failed to build Docker image")
	ErrRunDockerImage               = errors.New("failed to run Docker image")
	ErrDockerComposeNotImplemented  = errors.New("docker compose deployment not implemented yet")
	ErrMissingImageName             = errors.New("image name is required")
	ErrFailedToListContainers       = errors.New("failed to list containers")
	ErrFailedToCreateContainer      = errors.New("failed to create container")
	ErrFailedToStartNewContainer    = errors.New("failed to start new container")
	ErrFailedToUpdateContainer      = errors.New("failed to update container")
	ErrContainerNotRunning          = errors.New("container is not running")
	ErrDockerComposeFileNotFound    = errors.New("docker-compose file not found")
	ErrDockerComposeCommandFailed   = errors.New("docker-compose command failed")
	ErrDockerComposeInvalidConfig   = errors.New("invalid docker-compose configuration")
	ErrFailedToGetAvailablePort     = errors.New("failed to get available port")
)

const (
	LogDeploymentStarted                         = "Deployment started"
	LogRepositoryClonedSuccessfully              = "Repository cloned successfully"
	LogDeploymentCompletedSuccessfully           = "Deployment completed successfully"
	LogDockerImageBuiltSuccessfully              = "Docker image built successfully"
	LogStartingDockerImageBuild                  = "Starting Docker image build from Dockerfile"
	LogUsingDockerfileStrategy                   = "Using Dockerfile build strategy"
	LogUsingDockerComposeStrategy                = "Docker Compose deployment strategy selected"
	LogContainerRunning                          = "Container is running with ID: %s"
	LogApplicationExposed                        = "Application exposed on port: %d"
	LogBuildContextPath                          = "Build context path: %s"
	LogUsingBuildArgs                            = "Using %d build arguments"
	LogFailedToCreateApplicationRecord           = "Failed to create application record"
	LogFailedToCreateApplicationStatus           = "Failed to create application status: %s"
	LogFailedToCreateApplicationDeployment       = "Failed to create application deployment: %s"
	LogFailedToCreateApplicationDeploymentStatus = "Failed to create application deployment status: %s"
	LogFailedToCreateApplicationLogs             = "Failed to create application logs: %s"
	LogFailedToUpdateApplicationRecord           = "Failed to update application record"
	LogFailedToUpdateApplicationDeployment       = "Failed to update application deployment"
	LogFailedToParseRepositoryID                 = "Failed to parse repository ID: %s"
	LogFailedToCloneRepository                   = "Failed to clone repository: %s"
	LogFailedToCreateDeployment                  = "Failed to create deployment: %s"
	LogFailedToBuildDockerImage                  = "Failed to build Docker image: %s"
	LogFailedToRunDockerImage                    = "Failed to run Docker image: %s"
	LogDockerComposeNotImplemented               = "Docker compose deployment not implemented yet"
	LogDeploymentBuildPack                       = "Starting deployment process for build pack: %s"
	LogDockerComposeDeploymentStarted            = "Starting Docker Compose deployment"
	LogDockerComposeDeploymentCompleted          = "Docker Compose deployment completed successfully"
	LogDockerComposeDeploymentFailed             = "Docker Compose deployment failed: %s"
	LogRunningContainerFromImage                 = "Running container from image"
	LogPreparingToRunContainer                   = "Preparing to run container from image %s"
	LogEnvironmentVariables                      = "Environment variables: %v"
	LogContainerExposingPort                     = "Container will expose port %s"
	LogCreatingContainer                         = "Creating container..."
	LogContainerCreated                          = "Container created with ID: %s"
	LogStartingContainer                         = "Starting container..."
	LogContainerStartedSuccessfully              = "Container started successfully"
	LogFailedToCreateContainer                   = "Failed to create container: %s"
	LogFailedToStartContainer                    = "Failed to start container: %s"
	LogUpdatingContainer                         = "Updating container..."
	LogPreparingToUpdateContainer                = "Preparing to update container from image %s"
	LogFoundRunningContainer                     = "Found running container with ID: %s"
	LogNoRunningContainerFound                   = "No running container found"
	LogFailedToListContainers                    = "Failed to list containers: %s"
	LogFailedToUpdateContainer                   = "Failed to update container: %s"
	LogFailedToStopContainer                     = "Failed to stop container: %s"
	LogFailedToRemoveContainer                   = "Failed to remove container: %s"
	LogContainerStoppedSuccessfully              = "Container stopped successfully"
	LogStartingNewContainer                      = "Starting new container from image"
	LogCreatingNewContainer                      = "Creating new container..."
	LogNewContainerCreated                       = "New container created with ID"
	LogNewContainerStartedSuccessfully           = "New container started successfully"
	LogFailedToStopOldContainer                  = "Failed to stop old container: %s"
	LogRemovingOldContainer                      = "Removing old container..."
	LogOldContainerRemovedSuccessfully           = "Old container removed successfully"
	LogContainerUpdateCompleted                  = "Container update completed successfully"
	LogFailedToRemoveOldContainer                = "Failed to remove old container: %s"
	LogStoppingOldContainer                      = "Stopping old container..."
	LogRestartingContainer                       = "Restarting container..."
)
