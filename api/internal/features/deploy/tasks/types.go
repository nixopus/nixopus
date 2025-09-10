package tasks

import (
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	github_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type TaskService struct {
	Storage        storage.DeployRepository
	Logger         logger.Logger
	DockerRepo     docker.DockerRepository
	Github_service *github_service.GithubConnectorService
	Store          *shared_storage.Store
}

func NewTaskService(storage storage.DeployRepository, logger logger.Logger, dockerRepo docker.DockerRepository, githubService *github_service.GithubConnectorService, store *shared_storage.Store) *TaskService {
	return &TaskService{
		Storage:        storage,
		Logger:         logger,
		DockerRepo:     dockerRepo,
		Github_service: githubService,
		Store:          store,
	}
}
