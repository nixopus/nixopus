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
	DomainID             uuid.UUID                `json:"domain_id"`
	Environment          shared_types.Environment `json:"environment"`
	BuildPack            shared_types.BuildPack   `json:"build_pack"`
	Repository           string                   `json:"repository"`
	Branch               string                   `json:"branch"`
	PreRunCommand        string                   `json:"pre_run_command"`
	PostRunCommand       string                   `json:"post_run_command"`
	BuildVariables       map[string]string        `json:"build_variables"`
	EnvironmentVariables map[string]string        `json:"environment_variables"`
	Port                 int                      `json:"port"`
}

var (
	ErrInvalidRequestType           = errors.New("invalid request type")
	ErrMissingName                  = errors.New("name is required")
	ErrMissingDomainID              = errors.New("domain_id is required")
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
	LogFailedToParseRepositoryID                 = "Failed to parse repository ID: %s"
	LogFailedToCloneRepository                   = "Failed to clone repository: %s"
	LogFailedToCreateDeployment                  = "Failed to create deployment: %s"
	LogFailedToBuildDockerImage                  = "Failed to build Docker image: %s"
	LogFailedToRunDockerImage                    = "Failed to run Docker image: %s"
	LogDockerComposeNotImplemented               = "Docker compose deployment not implemented yet"
	LogDeploymentBuildPack                       = "Starting deployment process for build pack: %s"
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
)
