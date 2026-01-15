package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/service"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type HealthCheckController struct {
	store     *shared_storage.Store
	validator *validation.Validator
	service   *service.HealthCheckService
	ctx       context.Context
	logger    logger.Logger
}

func NewHealthCheckController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) *HealthCheckController {
	healthCheckStorage := storage.HealthCheckStorage{DB: store.DB, Ctx: ctx}
	healthCheckService := service.NewHealthCheckService(store, ctx, l, &healthCheckStorage)
	return &HealthCheckController{
		store:     store,
		validator: validation.NewValidator(&healthCheckStorage),
		service:   healthCheckService,
		ctx:       ctx,
		logger:    l,
	}
}
