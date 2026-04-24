package scheduler

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/config"
	healthcheck_service "github.com/nixopus/nixopus/api/internal/features/healthcheck/service"
	healthcheck_storage "github.com/nixopus/nixopus/api/internal/features/healthcheck/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/scheduler/billing"
	"github.com/nixopus/nixopus/api/internal/scheduler/cleanup"
	"github.com/nixopus/nixopus/api/internal/scheduler/container"
	"github.com/nixopus/nixopus/api/internal/scheduler/machines"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
)

type Schedulers struct {
	Main                *Scheduler
	HealthCheck         *HealthCheckScheduler
	Billing             *billing.Billing
	Backup              *machines.Backup
	TrialExpiry         *machines.TrialExpiry
	StaleMachineCleanup *machines.StaleCleanup
	MachineHealthCheck  *machines.HealthCheck
}

// InitSchedulers creates and configures all schedulers
func InitSchedulers(store *shared_storage.Store, ctx context.Context) *Schedulers {
	l := logger.NewLogger()

	sched := NewScheduler(store.DB, ctx, l, DefaultSchedulerConfig())

	sched.RegisterJob(cleanup.NewDeploymentLogsCleanupJob(store.DB, l))
	sched.RegisterJob(cleanup.NewAuditLogsCleanupJob(store.DB, l))

	sched.RegisterJob(container.NewPruneImagesJob(l))
	sched.RegisterJob(container.NewPruneBuildCacheJob(l))

	healthCheckStorage := healthcheck_storage.HealthCheckStorage{DB: store.DB, Ctx: ctx}
	healthCheckService := healthcheck_service.NewHealthCheckService(store, ctx, l, &healthCheckStorage)
	healthCheckScheduler := NewHealthCheckScheduler(healthCheckService, l, ctx, nil)

	billingScheduler := billing.New(store.DB, ctx, l)
	backupScheduler := machines.NewBackup(store.DB, ctx, l)
	trialExpiryScheduler := machines.NewTrialExpiry(store.DB, ctx, l, config.AppConfig.Trail.TrialPeriodDays)
	staleMachineCleanup := machines.NewStaleCleanup(store.DB, ctx, l)
	machineHealthCheck := machines.NewHealthCheck(store.DB, ctx, l)

	return &Schedulers{
		Main:                sched,
		HealthCheck:         healthCheckScheduler,
		Billing:             billingScheduler,
		Backup:              backupScheduler,
		TrialExpiry:         trialExpiryScheduler,
		StaleMachineCleanup: staleMachineCleanup,
		MachineHealthCheck:  machineHealthCheck,
	}
}
