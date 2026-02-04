package tasks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	github_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type TaskService struct {
	Storage        storage.DeployRepository
	Logger         logger.Logger
	Github_service *github_service.GithubConnectorService
	Store          *shared_storage.Store
}

func NewTaskService(storage storage.DeployRepository, logger logger.Logger, githubService *github_service.GithubConnectorService, store *shared_storage.Store) *TaskService {
	return &TaskService{
		Storage:        storage,
		Logger:         logger,
		Github_service: githubService,
		Store:          store,
	}
}

// getDockerService retrieves docker service from context (organization-aware)
func (s *TaskService) getDockerService(ctx context.Context) (docker.DockerRepository, error) {
	dockerService, err := docker.GetDockerServiceFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker service from context: %w", err)
	}
	if dockerService == nil {
		return nil, fmt.Errorf("docker service is nil")
	}
	return dockerService, nil
}

// LiveDevConfig holds configuration for starting a live dev service
type LiveDevConfig struct {
	// ApplicationID is the associated application ID for logging
	ApplicationID uuid.UUID

	// OrganizationID is the organization ID for organization-specific operations
	OrganizationID uuid.UUID

	// StagingPath is the local filesystem path containing the project files
	StagingPath string

	// Framework is the detected or specified framework (e.g., "nextjs", "vite")
	// If empty, auto-detection will be attempted
	Framework string

	// Port is the port to expose the dev server on (0 = auto-allocate)
	Port int

	// EnvVars are additional environment variables to set in the container
	EnvVars map[string]string

	// Domain is the domain name to route to this container (optional)
	Domain string
}
