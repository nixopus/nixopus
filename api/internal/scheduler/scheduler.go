package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
)

// Scheduler manages scheduled jobs for the application
type Scheduler struct {
	cron   *cron.Cron
	db     *bun.DB
	ctx    context.Context
	logger logger.Logger
	config SchedulerConfig
	jobs   []Job
	mu     sync.RWMutex
}

// cronLogger adapts our logger to cron's Logger interface
type cronLogger struct {
	l logger.Logger
}

func (c *cronLogger) Info(msg string, keysAndValues ...interface{}) {
	c.l.Log(logger.Info, msg, fmt.Sprint(keysAndValues...))
}

func (c *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	c.l.Log(logger.Error, fmt.Sprintf("%s: %v", msg, err), fmt.Sprint(keysAndValues...))
}

// NewScheduler creates a new scheduler instance
func NewScheduler(db *bun.DB, ctx context.Context, l logger.Logger, config SchedulerConfig) *Scheduler {
	cl := &cronLogger{l: l}
	return &Scheduler{
		cron: cron.New(
			cron.WithLogger(cl),
			cron.WithChain(
				cron.Recover(cl),
			),
		),
		db:     db,
		ctx:    ctx,
		logger: l,
		config: config,
		jobs:   make([]Job, 0),
	}
}

// RegisterJob adds a job to the scheduler
func (s *Scheduler) RegisterJob(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, job)
	s.logger.Log(logger.Info, fmt.Sprintf("Registered job: %s", job.Name()), "")
}

// Start begins the scheduler and runs all registered jobs on the configured schedule
func (s *Scheduler) Start() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, err := s.cron.AddFunc(s.config.Schedule, func() {
		s.runAllJobs()
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.cron.Start()
	s.logger.Log(logger.Info, fmt.Sprintf("Scheduler started with schedule: %s", s.config.Schedule), "")
	return nil
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() context.Context {
	s.logger.Log(logger.Info, "Stopping scheduler", "")
	return s.cron.Stop()
}

// RunNow immediately executes all jobs (useful for testing or manual triggers)
func (s *Scheduler) RunNow() {
	s.runAllJobs()
}

// runAllJobs executes all registered jobs for all organizations
func (s *Scheduler) runAllJobs() {
	s.logger.Log(logger.Info, "Starting scheduled job run", "")

	// Get all organizations with their settings (with timeout)
	queryCtx, queryCancel := context.WithTimeout(s.ctx, s.config.QueryTimeout)
	defer queryCancel()

	orgs, err := s.getOrganizationsWithSettings(queryCtx)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get organizations", err.Error())
		return
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Found %d organizations to process", len(orgs)), "")

	for _, org := range orgs {
		s.runJobsForOrganization(org)
	}

	s.logger.Log(logger.Info, "Completed scheduled job run", "")
}

// runJobsForOrganization executes all applicable jobs for a single organization
func (s *Scheduler) runJobsForOrganization(org *types.OrganizationSettings) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		if !job.IsEnabled(&org.Settings) {
			continue
		}

		// Create timeout context for this job
		jobCtx, jobCancel := context.WithTimeout(s.ctx, s.config.JobTimeout)

		start := time.Now()
		err := job.Run(jobCtx, org.OrganizationID)
		duration := time.Since(start)

		jobCancel() // Clean up context

		result := JobResult{
			JobName:        job.Name(),
			OrganizationID: org.OrganizationID,
			Success:        err == nil,
			Error:          err,
			ExecutedAt:     start,
			Duration:       duration,
		}

		s.logJobResult(result)
	}
}

// getOrganizationsWithSettings retrieves all organizations with their settings
func (s *Scheduler) getOrganizationsWithSettings(ctx context.Context) ([]*types.OrganizationSettings, error) {
	var settings []*types.OrganizationSettings
	err := s.db.NewSelect().
		Model(&settings).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// logJobResult logs the result of a job execution
func (s *Scheduler) logJobResult(result JobResult) {
	if result.Success {
		s.logger.Log(
			logger.Info,
			fmt.Sprintf("Job %s completed for org %s in %v", result.JobName, result.OrganizationID, result.Duration),
			"",
		)
	} else {
		s.logger.Log(
			logger.Error,
			fmt.Sprintf("Job %s failed for org %s: %v", result.JobName, result.OrganizationID, result.Error),
			"",
		)
	}
}
