package scheduler

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/scheduler/cleanup"
	"github.com/raghavyuva/nixopus-api/internal/scheduler/container"
	"github.com/uptrace/bun"
)

// InitScheduler creates and configures the scheduler with all jobs
func InitScheduler(db *bun.DB, ctx context.Context) *Scheduler {
	l := logger.NewLogger()
	sched := NewScheduler(db, ctx, l, DefaultSchedulerConfig())

	// Register cleanup jobs
	sched.RegisterJob(cleanup.NewDeploymentLogsCleanupJob(db, l))
	sched.RegisterJob(cleanup.NewAuditLogsCleanupJob(db, l))
	sched.RegisterJob(cleanup.NewExtensionLogsCleanupJob(db, l))

	// Register container maintenance jobs
	sched.RegisterJob(container.NewPruneImagesJob(l))
	sched.RegisterJob(container.NewPruneBuildCacheJob(l))

	return sched
}
