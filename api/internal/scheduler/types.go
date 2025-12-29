package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// Job defines the interface that all scheduled jobs must implement
type Job interface {
	// Name returns the unique name of the job
	Name() string

	// Run executes the job for a specific organization
	Run(ctx context.Context, orgID uuid.UUID) error

	// IsEnabled checks if the job is enabled for the given organization settings
	IsEnabled(settings *types.OrganizationSettingsData) bool

	// GetRetentionDays returns the retention period in days from settings
	GetRetentionDays(settings *types.OrganizationSettingsData) int
}

// JobResult represents the result of a job execution
type JobResult struct {
	JobName        string
	OrganizationID uuid.UUID
	Success        bool
	RecordsDeleted int64
	Error          error
	ExecutedAt     time.Time
	Duration       time.Duration
}

// SchedulerConfig holds configuration for the scheduler
type SchedulerConfig struct {
	// Schedule is the cron expression for when jobs run (default: "0 2 * * *" = 2 AM daily)
	Schedule string

	// DefaultRetentionDays is the fallback retention period if not set in org settings
	DefaultRetentionDays int

	// JobTimeout is the maximum duration a single job can run before being cancelled
	JobTimeout time.Duration

	// QueryTimeout is the maximum duration for database queries
	QueryTimeout time.Duration
}

// DefaultSchedulerConfig returns the default scheduler configuration
func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		Schedule:             "0 2 * * *", // 2 AM daily
		DefaultRetentionDays: 30,
		JobTimeout:           5 * time.Minute,  // Max 5 minutes per job per org
		QueryTimeout:         30 * time.Second, // Max 30 seconds for DB queries
	}
}
