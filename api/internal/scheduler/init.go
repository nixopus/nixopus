package scheduler

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/scheduler/cleanup"
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

	return sched
}
