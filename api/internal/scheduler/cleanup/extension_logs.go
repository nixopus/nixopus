package cleanup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

const (
	ExtensionLogsCleanupJobName   = "extension_logs_cleanup"
	DefaultExtensionLogsRetention = 30
)

// ExtensionLogsCleanupJob handles cleanup of old extension logs
// Note: Extension logs are system-wide (not org-scoped), so this job
// uses the maximum retention period across all organizations to ensure
// logs are only deleted when ALL organizations agree they should be deleted.
// A mutex ensures the cleanup runs only once per scheduler cycle.
type ExtensionLogsCleanupJob struct {
	db            *bun.DB
	logger        logger.Logger
	mu            sync.Mutex
	lastCleanupAt time.Time
}

// NewExtensionLogsCleanupJob creates a new extension logs cleanup job
func NewExtensionLogsCleanupJob(db *bun.DB, l logger.Logger) *ExtensionLogsCleanupJob {
	return &ExtensionLogsCleanupJob{
		db:     db,
		logger: l,
	}
}

// Name returns the job name
func (j *ExtensionLogsCleanupJob) Name() string {
	return ExtensionLogsCleanupJobName
}

// IsEnabled checks if cleanup is enabled for this organization
func (j *ExtensionLogsCleanupJob) IsEnabled(settings *types.OrganizationSettingsData) bool {
	return IsEnabledOrDefault(settings.ExtensionLogsCleanupEnabled)
}

// GetRetentionDays returns the retention period from settings
func (j *ExtensionLogsCleanupJob) GetRetentionDays(settings *types.OrganizationSettingsData) int {
	return GetRetentionDaysOrDefault(settings.ExtensionLogsRetentionDays, DefaultExtensionLogsRetention)
}

// Run executes the cleanup job
// Since extension logs are system-wide (not org-scoped), this job:
// 1. Ensures cleanup runs only once per scheduler cycle using a mutex
// 2. Queries the maximum retention period across all organizations
// 3. Uses that maximum to ensure logs are only deleted when ALL orgs agree
func (j *ExtensionLogsCleanupJob) Run(ctx context.Context, orgID uuid.UUID) error {
	// Use mutex to ensure only one cleanup runs per scheduler cycle
	// Check if we've already run cleanup within the last minute (scheduler cycle protection)
	j.mu.Lock()
	if time.Since(j.lastCleanupAt) < time.Minute {
		j.mu.Unlock()
		j.logger.Log(
			logger.Info,
			fmt.Sprintf("Extension logs cleanup already ran recently, skipping for org %s", orgID),
			"",
		)
		return nil
	}
	j.lastCleanupAt = time.Now()
	j.mu.Unlock()

	// Query the maximum retention period across all organizations
	// This ensures we only delete logs that ALL orgs agree should be deleted
	var maxRetentionResult struct {
		MaxRetention int `bun:"max_retention"`
	}
	err := j.db.NewSelect().
		TableExpr("organization_settings").
		ColumnExpr("COALESCE(MAX((settings->>'extension_logs_retention_days')::int), ?) AS max_retention", DefaultExtensionLogsRetention).
		Scan(ctx, &maxRetentionResult)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No settings found, use default
			maxRetentionResult.MaxRetention = DefaultExtensionLogsRetention
		} else {
			return fmt.Errorf("failed to get max retention period: %w", err)
		}
	}

	retentionDays := maxRetentionResult.MaxRetention
	if retentionDays <= 0 {
		retentionDays = DefaultExtensionLogsRetention
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Running extension logs cleanup (max retention across all orgs: %d days, cutoff: %s)",
			retentionDays, cutoffDate.Format(time.RFC3339)),
		"",
	)

	// Delete old extension logs (system-wide)
	result, err := j.db.NewDelete().
		TableExpr("extension_logs").
		Where("created_at < ?", cutoffDate).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete extension logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Deleted %d extension logs (system-wide cleanup with max retention %d days)", rowsAffected, retentionDays),
		"",
	)

	return nil
}
