package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// DeployService handles read operations and validation for deployments.
// Deployment creation, update, rollback, restart, and delete operations
// are handled by TaskService in the tasks package.
type DeployService struct {
	storage storage.DeployRepository
	Ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewDeployService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, deploy_repo storage.DeployRepository) *DeployService {
	return &DeployService{
		storage: deploy_repo,
		store:   store,
		Ctx:     ctx,
		logger:  logger,
	}
}
