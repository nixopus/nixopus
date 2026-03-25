package service

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/healthcheck/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
)

type HealthCheckService struct {
	storage storage.HealthCheckRepository
	store   *shared_storage.Store
	ctx     context.Context
	logger  logger.Logger
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
