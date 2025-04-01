package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	github_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type DeployService struct {
	storage        storage.DeployRepository
	Ctx            context.Context
	store          *shared_storage.Store
	logger         logger.Logger
	dockerRepo     docker.DockerRepository
	github_service *github_service.GithubConnectorService
}

func NewDeployService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, deploy_repo storage.DeployRepository, dockerRepo docker.DockerRepository, github_service *github_service.GithubConnectorService) *DeployService {
	return &DeployService{
		storage:        deploy_repo,
		store:          store,
		Ctx:            ctx,
		logger:         logger,
		dockerRepo:     dockerRepo,
		github_service: github_service,
	}
}

func (s *DeployService) GetApplicationDeployments(applicationID uuid.UUID, page, pageSize int) ([]shared_types.ApplicationDeployment, int, error) {
	return s.storage.GetPaginatedApplicationDeployments(applicationID, page, pageSize)
}
