package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// ApplicationProvider is a minimal interface for fetching application data
// during health checks. It is satisfied by deploy/storage.DeployStorage in
// production and by a test double in unit tests.
type ApplicationProvider interface {
	GetApplicationById(id string, organizationID uuid.UUID) (shared_types.Application, error)
	GetApplicationDomains(applicationID uuid.UUID) ([]shared_types.ApplicationDomain, error)
}

type HealthCheckService struct {
	storage     storage.HealthCheckRepository
	store       *shared_storage.Store
	ctx         context.Context
	logger      logger.Logger
	appProvider ApplicationProvider // nil → lazy-init from store in execute_check.go
}

func NewHealthCheckService(
	store *shared_storage.Store,
	ctx context.Context,
	logger logger.Logger,
	healthCheckRepo storage.HealthCheckRepository,
) *HealthCheckService {
	return &HealthCheckService{
		storage: healthCheckRepo,
		store:   store,
		ctx:     ctx,
		logger:  logger,
	}
}
