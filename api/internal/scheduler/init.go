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
	Billing     *BillingScheduler
}

// InitSchedulers creates and configures all schedulers
func InitSchedulers(store *shared_storage.Store, ctx context.Context) *Schedulers {
	l := logger.NewLogger()

	sched := NewScheduler(store.DB, ctx, l, DefaultSchedulerConfig())

	sched.RegisterJob(cleanup.NewDeploymentLogsCleanupJob(store.DB, l))
	sched.RegisterJob(cleanup.NewAuditLogsCleanupJob(store.DB, l))
	sched.RegisterJob(cleanup.NewExtensionLogsCleanupJob(store.DB, l))

	sched.RegisterJob(container.NewPruneImagesJob(l))
	sched.RegisterJob(container.NewPruneBuildCacheJob(l))

	healthCheckStorage := healthcheck_storage.HealthCheckStorage{DB: store.DB, Ctx: ctx}
	healthCheckService := healthcheck_service.NewHealthCheckService(store, ctx, l, &healthCheckStorage)
	healthCheckScheduler := NewHealthCheckScheduler(healthCheckService, l, ctx, nil)

	billingScheduler := NewBillingScheduler(store.DB, ctx, l)

	return &Schedulers{
		Main:        sched,
		HealthCheck: healthCheckScheduler,
		Billing:     billingScheduler,
	}
}
