package types

import (
	"errors"

	"github.com/google/uuid"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type IsNameAlreadyTakenRequest struct {
	Name string `json:"name" validate:"required" description:"Application name to check for uniqueness" example:"my-web-app"`
}

type IsDomainAlreadyTakenRequest struct {
	Domain string `json:"domain" validate:"required" description:"Domain to check for availability" example:"app.example.com"`
}

type IsDomainValidRequest struct {
	Domain string `json:"domain" validate:"required" description:"Domain to validate" example:"app.example.com"`
}

type IsPortAlreadyTakenRequest struct {
	Port int `json:"port" validate:"required" description:"Port number to check for availability" example:"3000"`
}

// ComposeDomain maps a domain to a specific compose service or port override.
type ComposeDomain struct {
	Domain      string `json:"domain" validate:"required" description:"Domain name for the compose service" example:"app.example.com"`
	ServiceName string `json:"service_name,omitempty" description:"Name of the compose service to route to" example:"web"`
	Port        int    `json:"port,omitempty" description:"Port override for the compose service" example:"8080"`
}

type CreateDeploymentRequest struct {
	Name                 string                       `json:"name" validate:"required" description:"Application name" example:"my-web-app"`
	Domains              []string                     `json:"domains,omitempty" validate:"omitempty,max=5" description:"Custom domains for the application"`
	ComposeDomains       []ComposeDomain              `json:"compose_domains,omitempty" description:"Domain-to-service mappings for compose deployments"`
	Environment          shared_types.Environment     `json:"environment" validate:"required" description:"Deployment environment" example:"production"`
	BuildPack            shared_types.BuildPack       `json:"build_pack" validate:"required" description:"Build strategy for the application" example:"dockerfile"`
	Repository           string                       `json:"repository" validate:"required" description:"Git repository URL or identifier" example:"github.com/org/repo"`
	Branch               string                       `json:"branch" validate:"required" description:"Git branch to deploy" example:"main"`
	PreRunCommand        string                       `json:"pre_run_command" description:"Command to run before the application starts" example:"npm run migrate"`
	PostRunCommand       string                       `json:"post_run_command" description:"Command to run after the application starts" example:"npm run seed"`
	BuildVariables       map[string]string            `json:"build_variables" description:"Key-value pairs passed as build arguments to Docker"`
	EnvironmentVariables map[string]string            `json:"environment_variables" description:"Key-value pairs set as runtime environment variables"`
	Port                 int                          `json:"port" validate:"required,min=1,max=65535" description:"Port the application listens on" example:"3000"`
	DockerfilePath       string                       `json:"dockerfile_path,omitempty" description:"Path to the Dockerfile relative to the repository root" example:"Dockerfile"`
	BasePath             string                       `json:"base_path,omitempty" description:"Base path for the application. Defaults to /" example:"/"`
	Source               shared_types.Source          `json:"source,omitempty" description:"Source type of the repository" example:"github"`
	ServerIDs            []uuid.UUID                  `json:"server_ids,omitempty" description:"Server IDs to deploy the application to"`
	PrimaryServerID      *uuid.UUID                   `json:"primary_server_id,omitempty" description:"Primary server ID for routing" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoutingStrategy      shared_types.RoutingStrategy `json:"routing_strategy,omitempty" description:"Routing strategy for multi-server deployments" example:"round-robin"`
	TargetServerIDs      []uuid.UUID                  `json:"target_server_ids,omitempty" description:"Specific server IDs to target for this deployment"`
}

// CreateProjectRequest is used to create a project (application) without triggering deployment.
type CreateProjectRequest struct {
	Name                 string                       `json:"name" validate:"required" description:"Project name" example:"my-web-app"`
	Domains              []string                     `json:"domains,omitempty" description:"Custom domains for the project"`
	ComposeDomains       []ComposeDomain              `json:"compose_domains,omitempty" description:"Domain-to-service mappings for compose deployments"`
	ComposeServices      []PreviewComposeService      `json:"compose_services,omitempty" description:"Compose services configuration"`
	Environment          shared_types.Environment     `json:"environment,omitempty" description:"Deployment environment. Defaults to production if not specified" example:"production"`
	BuildPack            shared_types.BuildPack       `json:"build_pack,omitempty" description:"Build strategy. Defaults to dockerfile if not specified" example:"dockerfile"`
	Repository           string                       `json:"repository" validate:"required" description:"Git repository URL or identifier" example:"github.com/org/repo"`
	Branch               string                       `json:"branch,omitempty" description:"Git branch to deploy. Defaults to main if not specified" example:"main"`
	PreRunCommand        string                       `json:"pre_run_command,omitempty" description:"Command to run before the application starts" example:"npm run migrate"`
	PostRunCommand       string                       `json:"post_run_command,omitempty" description:"Command to run after the application starts" example:"npm run seed"`
	BuildVariables       map[string]string            `json:"build_variables,omitempty" description:"Key-value pairs passed as build arguments to Docker"`
	EnvironmentVariables map[string]string            `json:"environment_variables,omitempty" description:"Key-value pairs set as runtime environment variables"`
	Port                 int                          `json:"port,omitempty" description:"Port the application listens on. Defaults to 3000 if not specified" example:"3000"`
	DockerfilePath       string                       `json:"dockerfile_path,omitempty" description:"Path to the Dockerfile. Defaults to Dockerfile if not specified" example:"Dockerfile"`
	BasePath             string                       `json:"base_path,omitempty" description:"Base path for the application. Defaults to / if not specified" example:"/"`
	Source               shared_types.Source          `json:"source,omitempty" description:"Source type of the repository" example:"github"`
	ServerIDs            []uuid.UUID                  `json:"server_ids,omitempty" description:"Server IDs to deploy the application to"`
	PrimaryServerID      *uuid.UUID                   `json:"primary_server_id,omitempty" description:"Primary server ID for routing" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoutingStrategy      shared_types.RoutingStrategy `json:"routing_strategy,omitempty" description:"Routing strategy for multi-server deployments" example:"round-robin"`
}

type PreviewComposeRequest struct {
	Repository     string `json:"repository" validate:"required" description:"Git repository URL or identifier" example:"github.com/org/repo"`
	Branch         string `json:"branch" validate:"required" description:"Git branch to preview" example:"main"`
	BasePath       string `json:"base_path,omitempty" description:"Base path within the repository" example:"/"`
	DockerfilePath string `json:"dockerfile_path,omitempty" description:"Path to the Dockerfile" example:"Dockerfile"`
}

type PreviewComposeService struct {
	ServiceName string `json:"service_name" validate:"required" description:"Name of the compose service" example:"web"`
	Port        int    `json:"port" validate:"required" description:"Port the service listens on" example:"8080"`
}

type PreviewComposeResponse struct {
	Services []PreviewComposeService `json:"services"`
}

// DeployProjectRequest is used to trigger deployment of an existing project (application).
type DeployProjectRequest struct {
	ID uuid.UUID `json:"id" validate:"required" description:"Project ID to deploy" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UpdateDeploymentRequest struct {
	Name                 string                       `json:"name,omitempty" validate:"omitempty,min=3" description:"Application name. Minimum 3 characters if provided" example:"my-web-app"`
	Environment          shared_types.Environment     `json:"environment,omitempty" validate:"omitempty" description:"Deployment environment. Must be valid if provided" example:"production"`
	BuildPack            shared_types.BuildPack       `json:"build_pack,omitempty" validate:"omitempty" description:"Build strategy. Must be valid if provided" example:"dockerfile"`
	PreRunCommand        string                       `json:"pre_run_command,omitempty" description:"Command to run before the application starts" example:"npm run migrate"`
	PostRunCommand       string                       `json:"post_run_command,omitempty" description:"Command to run after the application starts" example:"npm run seed"`
	BuildVariables       map[string]string            `json:"build_variables,omitempty" description:"Key-value pairs passed as build arguments to Docker"`
	EnvironmentVariables map[string]string            `json:"environment_variables,omitempty" description:"Key-value pairs set as runtime environment variables"`
	Port                 int                          `json:"port,omitempty" validate:"omitempty,min=1,max=65535" description:"Port the application listens on" example:"3000"`
	ID                   uuid.UUID                    `json:"id,omitempty" description:"Application ID to update" example:"550e8400-e29b-41d4-a716-446655440000"`
	Force                bool                         `json:"force,omitempty" description:"Force the update even if a deployment is in progress"`
	DockerfilePath       string                       `json:"dockerfile_path,omitempty" description:"Path to the Dockerfile relative to the repository root" example:"Dockerfile"`
	BasePath             string                       `json:"base_path,omitempty" description:"Base path for the application" example:"/"`
	Domains              []string                     `json:"domains,omitempty" validate:"omitempty,max=5" description:"Custom domains for the application. Maximum 5 allowed"`
	ComposeDomains       []ComposeDomain              `json:"compose_domains,omitempty" description:"Domain-to-service mappings for compose deployments"`
	RoutingStrategy      shared_types.RoutingStrategy `json:"routing_strategy,omitempty" description:"Routing strategy for multi-server deployments" example:"round-robin"`
}

type DeleteDeploymentRequest struct {
	ID uuid.UUID `json:"id" validate:"required" description:"Application ID to delete" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type ReDeployApplicationRequest struct {
	ID                uuid.UUID   `json:"id" validate:"required" description:"Application ID to redeploy" example:"550e8400-e29b-41d4-a716-446655440000"`
	Force             bool        `json:"force" description:"Force redeployment even if already deploying"`
	ForceWithoutCache bool        `json:"force_without_cache" description:"Force redeployment without using Docker build cache"`
	TargetServerIDs   []uuid.UUID `json:"target_server_ids,omitempty" description:"Specific server IDs to target for this redeployment"`
}

type RollbackDeploymentRequest struct {
	ID              uuid.UUID   `json:"id" validate:"required" description:"Application ID to roll back" example:"550e8400-e29b-41d4-a716-446655440000"`
	TargetServerIDs []uuid.UUID `json:"target_server_ids,omitempty" description:"Specific server IDs to target for the rollback"`
}

type RestartDeploymentRequest struct {
	ID uuid.UUID `json:"id" validate:"required" description:"Application ID to restart" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// DuplicateProjectRequest is used to create a duplicate of an existing project with a different environment.
type DuplicateProjectRequest struct {
	SourceProjectID uuid.UUID                    `json:"source_project_id" validate:"required" description:"ID of the project to duplicate" example:"550e8400-e29b-41d4-a716-446655440000"`
	Domains         []string                     `json:"domains,omitempty" description:"Custom domains for the duplicated project"`
	Environment     shared_types.Environment     `json:"environment" validate:"required" description:"Environment for the duplicated project" example:"staging"`
	Branch          string                       `json:"branch,omitempty" description:"Git branch override for the duplicate" example:"develop"`
	ServerIDs       []uuid.UUID                  `json:"server_ids,omitempty" description:"Server IDs to deploy the duplicate to"`
	PrimaryServerID *uuid.UUID                   `json:"primary_server_id,omitempty" description:"Primary server ID for routing" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoutingStrategy shared_types.RoutingStrategy `json:"routing_strategy,omitempty" description:"Routing strategy for multi-server deployments" example:"round-robin"`
}

// GetProjectFamilyRequest is used to get all projects in a family.
type GetProjectFamilyRequest struct {
	FamilyID uuid.UUID `json:"family_id" validate:"required" description:"Project family ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// SetApplicationServersRequest is used to assign servers to an existing application.
type SetApplicationServersRequest struct {
	ApplicationID   uuid.UUID                    `json:"application_id" validate:"required" description:"Application ID to assign servers to" example:"550e8400-e29b-41d4-a716-446655440000"`
	ServerIDs       []uuid.UUID                  `json:"server_ids" validate:"required" description:"Server IDs to assign to the application"`
	PrimaryServerID *uuid.UUID                   `json:"primary_server_id,omitempty" description:"Primary server ID for routing" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoutingStrategy shared_types.RoutingStrategy `json:"routing_strategy,omitempty" description:"Routing strategy for multi-server deployments" example:"round-robin"`
}

// AddApplicationToFamilyRequest is used to add a new application to an existing family.
type AddApplicationToFamilyRequest struct {
	FamilyID             *uuid.UUID               `json:"family_id,omitempty" description:"Family ID to add the application to. Creates new family if not provided" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name                 string                   `json:"name" validate:"required" description:"Application name" example:"my-api"`
	Path                 string                   `json:"path" description:"Base path for the application. Defaults to /" example:"api/"`
	Repository           string                   `json:"repository" validate:"required" description:"Git repository URL or identifier" example:"github.com/org/repo"`
	Branch               string                   `json:"branch,omitempty" description:"Git branch. Defaults to main if not specified" example:"main"`
	Environment          shared_types.Environment `json:"environment,omitempty" description:"Deployment environment. Defaults to development if not specified" example:"development"`
	BuildPack            shared_types.BuildPack   `json:"build_pack,omitempty" description:"Build strategy. Defaults to dockerfile if not specified" example:"dockerfile"`
	Port                 int                      `json:"port,omitempty" description:"Port the application listens on. Defaults to 8080 if not specified" example:"8080"`
	DockerfilePath       string                   `json:"dockerfile_path,omitempty" description:"Path to the Dockerfile. Defaults to Dockerfile if not specified" example:"Dockerfile"`
	PreRunCommand        string                   `json:"pre_run_command,omitempty" description:"Command to run before the application starts" example:"npm run migrate"`
	PostRunCommand       string                   `json:"post_run_command,omitempty" description:"Command to run after the application starts" example:"npm run seed"`
	BuildVariables       map[string]string        `json:"build_variables,omitempty" description:"Key-value pairs passed as build arguments to Docker"`
	EnvironmentVariables map[string]string        `json:"environment_variables,omitempty" description:"Key-value pairs set as runtime environment variables"`
	Domains              []string                 `json:"domains,omitempty" description:"Custom domains for the application"`
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

// ComposeServicesResponse is the typed response for compose service listing.
type ComposeServicesResponse struct {
	Status  string                        `json:"status"`
	Message string                        `json:"message"`
	Data    []shared_types.ComposeService `json:"data"`
}

// ApplicationServersResponse is the typed response for application server assignment operations.
type ApplicationServersResponse struct {
	Status  string                           `json:"status"`
	Message string                           `json:"message"`
	Data    []shared_types.ApplicationServer `json:"data"`
}

type CreateTemplateDeploymentRequest struct {
	TemplateID      string                       `json:"template_id" validate:"required" description:"Template ID to deploy from" example:"wordpress"`
	Name            string                       `json:"name" validate:"required" description:"Name for the deployed application" example:"my-wordpress"`
	Variables       map[string]interface{}       `json:"variables" description:"Template-specific configuration variables"`
	ServerIDs       []uuid.UUID                  `json:"server_ids,omitempty" description:"Server IDs to deploy to"`
	PrimaryServerID *uuid.UUID                   `json:"primary_server_id,omitempty" description:"Primary server ID for routing" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoutingStrategy shared_types.RoutingStrategy `json:"routing_strategy,omitempty" description:"Routing strategy for multi-server deployments" example:"round-robin"`
	Environment     shared_types.Environment     `json:"environment,omitempty" description:"Deployment environment" example:"production"`
}

type CancelDeploymentRequest struct {
	DeploymentID uuid.UUID `json:"deployment_id" validate:"required" description:"Deployment ID to cancel" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type RecoverRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty" description:"Application ID to recover. If not provided, recovers all applications" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type RecoverAppResult struct {
	ApplicationID   uuid.UUID `json:"application_id"`
	ApplicationName string    `json:"application_name"`
	Reason          string    `json:"reason"`
}

type RecoverResult struct {
	Recovered []RecoverAppResult `json:"recovered"`
	Skipped   []RecoverAppResult `json:"skipped"`
	Failed    []RecoverAppResult `json:"failed"`
}

type RecoverResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Data    RecoverResult `json:"data"`
}

type IndexCodebaseResponse struct {
	Status  string                    `json:"status"`
	Message string                    `json:"message"`
	Data    IndexCodebaseResponseData `json:"data"`
}

type IndexCodebaseResponseData struct {
	Indexed int `json:"indexed"`
	Skipped int `json:"skipped"`
}

// Artifact represents a deployment's S3 image artifact metadata.
type Artifact struct {
	DeploymentID  string `json:"deployment_id"`
	ApplicationID string `json:"application_id"`
	AppName       string `json:"app_name"`
	S3Key         string `json:"s3_key"`
	Size          int64  `json:"size"`
	CreatedAt     string `json:"created_at"`
}

type ArtifactListResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Data    []Artifact `json:"data"`
}

type ArtifactDownloadResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		URL       string `json:"url"`
		ExpiresIn int    `json:"expires_in_seconds"`
	} `json:"data"`
}

type ArtifactDeleteResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	ErrAtLeastOneServerRequired         = errors.New("at least one server is required")
	ErrS3NotConfigured                  = errors.New("S3 image storage is not configured")
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
	ErrPaymentRequired                  = errors.New("payment required: deployment limit reached, please upgrade your plan")
	ErrDeploymentNotCancellable         = errors.New("deployment is not in a cancellable state")
	ErrDeploymentNotRunning             = errors.New("deployment not found or not running on this instance")
	ErrPermissionDenied                 = errors.New("permission denied")
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
