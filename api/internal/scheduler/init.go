package scheduler

import (
	"context"

	healthcheck_service "github.com/raghavyuva/nixopus-api/internal/features/healthcheck/service"
	healthcheck_storage "github.com/raghavyuva/nixopus-api/internal/features/healthcheck/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/scheduler/cleanup"
	"github.com/raghavyuva/nixopus-api/internal/scheduler/container"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type Schedulers struct {
	Main        *Scheduler
	HealthCheck *HealthCheckScheduler
}

// InitSchedulers creates and configures all schedulers
func InitSchedulers(store *shared_storage.Store, ctx context.Context) *Schedulers {
	l := logger.NewLogger()

	// Initialize main scheduler
	sched := NewScheduler(store.DB, ctx, l, DefaultSchedulerConfig())

	// Register cleanup jobs
	sched.RegisterJob(cleanup.NewDeploymentLogsCleanupJob(store.DB, l))
	sched.RegisterJob(cleanup.NewAuditLogsCleanupJob(store.DB, l))
	sched.RegisterJob(cleanup.NewExtensionLogsCleanupJob(store.DB, l))

	// Register container maintenance jobs
	sched.RegisterJob(container.NewPruneImagesJob(l))
	sched.RegisterJob(container.NewPruneBuildCacheJob(l))

	// Initialize health check scheduler (SocketServer will be set later from routes)
	healthCheckStorage := healthcheck_storage.HealthCheckStorage{DB: store.DB, Ctx: ctx}
	healthCheckService := healthcheck_service.NewHealthCheckService(store, ctx, l, &healthCheckStorage)
	healthCheckScheduler := NewHealthCheckScheduler(healthCheckService, l, ctx)

	return &Schedulers{
		Main:        sched,
		HealthCheck: healthCheckScheduler,
	}
}
