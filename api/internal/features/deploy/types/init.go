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
	Domains              []string                 `json:"domains,omitempty"`
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

// CreateProjectRequest is used to create a project (application) without triggering deployment.
type CreateProjectRequest struct {
	Name                 string                   `json:"name"`
	Domains              []string                 `json:"domains,omitempty"`
	Environment          shared_types.Environment `json:"environment,omitempty"`
	BuildPack            shared_types.BuildPack   `json:"build_pack,omitempty"`
	Repository           string                   `json:"repository"`
	Branch               string                   `json:"branch,omitempty"`
	PreRunCommand        string                   `json:"pre_run_command,omitempty"`
	PostRunCommand       string                   `json:"post_run_command,omitempty"`
	BuildVariables       map[string]string        `json:"build_variables,omitempty"`
	EnvironmentVariables map[string]string        `json:"environment_variables,omitempty"`
	Port                 int                      `json:"port,omitempty"`
	DockerfilePath       string                   `json:"dockerfile_path,omitempty"`
	BasePath             string                   `json:"base_path,omitempty"`
}

// DeployProjectRequest is used to trigger deployment of an existing project (application).
type DeployProjectRequest struct {
	ID uuid.UUID `json:"id"`
}

type UpdateDeploymentRequest struct {
	Name                 string                   `json:"name,omitempty"`
	Environment          shared_types.Environment `json:"environment,omitempty"`
	PreRunCommand        string                   `json:"pre_run_command,omitempty"`
	PostRunCommand       string                   `json:"post_run_command,omitempty"`
	BuildVariables       map[string]string        `json:"build_variables,omitempty"`
	EnvironmentVariables map[string]string        `json:"environment_variables,omitempty"`
	Port                 int                      `json:"port,omitempty"`
	ID                   uuid.UUID                `json:"id,omitempty"`
	Force                bool                     `json:"force,omitempty"`
	DockerfilePath       string                   `json:"dockerfile_path,omitempty"`
	BasePath             string                   `json:"base_path,omitempty"`
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

// DuplicateProjectRequest is used to create a duplicate of an existing project with a different environment.
type DuplicateProjectRequest struct {
	SourceProjectID uuid.UUID                `json:"source_project_id"`
	Domains         []string                 `json:"domains,omitempty"`
	Environment     shared_types.Environment `json:"environment"`
	Branch          string                   `json:"branch,omitempty"`
}

// GetProjectFamilyRequest is used to get all projects in a family.
type GetProjectFamilyRequest struct {
	FamilyID uuid.UUID `json:"family_id"`
}

// ProjectFamilyResponseData contains the data for project family response.
type ProjectFamilyResponseData struct {
	Projects []shared_types.Application `json:"projects"`
}

// ProjectFamilyResponse is the typed response for project family.
type ProjectFamilyResponse struct {
	Status  string                    `json:"status"`
	Message string                    `json:"message"`
	Data    ProjectFamilyResponseData `json:"data"`
}

// EnvironmentsInFamilyResponseData contains the environments in a family.
type EnvironmentsInFamilyResponseData struct {
	Environments []shared_types.Environment `json:"environments"`
}

// EnvironmentsInFamilyResponse is the typed response for environments in family.
type EnvironmentsInFamilyResponse struct {
	Status  string                           `json:"status"`
	Message string                           `json:"message"`
	Data    EnvironmentsInFamilyResponseData `json:"data"`
}

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ApplicationResponse is the typed response for single application operations
type ApplicationResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Data    shared_types.Application `json:"data"`
}

// ListApplicationsResponseData contains the data for list applications response
type ListApplicationsResponseData struct {
	Applications []shared_types.Application `json:"applications"`
	TotalCount   int                        `json:"total_count"`
	Page         string                     `json:"page"`
	PageSize     string                     `json:"page_size"`
}

// ListApplicationsResponse is the typed response for listing applications
type ListApplicationsResponse struct {
	Status  string                       `json:"status"`
	Message string                       `json:"message"`
	Data    ListApplicationsResponseData `json:"data"`
}

// DeploymentResponse is the typed response for single deployment
type DeploymentResponse struct {
	Status  string                             `json:"status"`
	Message string                             `json:"message"`
	Data    shared_types.ApplicationDeployment `json:"data"`
}

// ListDeploymentsResponseData contains the data for list deployments response
type ListDeploymentsResponseData struct {
	Deployments []shared_types.ApplicationDeployment `json:"deployments"`
	TotalCount  int                                  `json:"total_count"`
	Page        string                               `json:"page"`
	PageSize    string                               `json:"page_size"`
}

// ListDeploymentsResponse is the typed response for listing deployments
type ListDeploymentsResponse struct {
	Status  string                      `json:"status"`
	Message string                      `json:"message"`
	Data    ListDeploymentsResponseData `json:"data"`
}

// LogsResponseData contains the data for logs response
type LogsResponseData struct {
	Logs       []shared_types.ApplicationLogs `json:"logs"`
	TotalCount int64                          `json:"total_count"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
}

// LogsResponse is the typed response for logs
type LogsResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Data    LogsResponseData `json:"data"`
}

// LabelsResponse is the typed response for labels update
type LabelsResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

var (
	ErrMissingID                        = errors.New("id is required")
	ErrInvalidRequestType               = errors.New("invalid request type")
	ErrMissingName                      = errors.New("name is required")
	ErrMissingDomain                    = errors.New("domain is required")
	ErrMissingRepository                = errors.New("repository is required")
	ErrMissingBranch                    = errors.New("branch is required")
	ErrMissingPort                      = errors.New("port is required")
	ErrInvalidEnvironment               = errors.New("invalid environment")
	ErrInvalidBuildPack                 = errors.New("invalid build pack")
	ErrFailedToCreateTarFromContext     = errors.New("failed to create tar from context")
	ErrProcessingBuildOutput            = errors.New("failed to process build output")
	ErrBuildDockerImage                 = errors.New("failed to build Docker image")
	ErrRunDockerImage                   = errors.New("failed to run Docker image")
	ErrDockerComposeNotImplemented      = errors.New("docker compose deployment not implemented yet")
	ErrMissingImageName                 = errors.New("image name is required")
	ErrFailedToListContainers           = errors.New("failed to list containers")
	ErrFailedToCreateContainer          = errors.New("failed to create container")
	ErrFailedToStartNewContainer        = errors.New("failed to start new container")
	ErrFailedToUpdateContainer          = errors.New("failed to update container")
	ErrContainerNotRunning              = errors.New("container is not running")
	ErrDockerComposeFileNotFound        = errors.New("docker-compose file not found")
	ErrDockerComposeCommandFailed       = errors.New("docker-compose command failed")
	ErrDockerComposeInvalidConfig       = errors.New("invalid docker-compose configuration")
	ErrFailedToGetAvailablePort         = errors.New("failed to get available port")
	ErrApplicationNotFound              = errors.New("application not found")
	ErrApplicationNotDraft              = errors.New("application is not in draft status, cannot deploy")
	ErrApplicationAlreadyDeployed       = errors.New("application has already been deployed")
	ErrMissingSourceProjectID           = errors.New("source project id is required")
	ErrEnvironmentAlreadyExistsInFamily = errors.New("a project with this environment already exists in the family")
	ErrSameEnvironmentAsDuplicate       = errors.New("cannot duplicate project with the same environment")
	ErrProjectFamilyNotFound            = errors.New("project family not found")
	ErrDomainLimitReached               = errors.New("maximum of 5 domains per application reached")
	ErrDomainAlreadyExists              = errors.New("domain already exists for this application")
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
