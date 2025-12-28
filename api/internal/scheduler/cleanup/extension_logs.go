package cleanup

import (
	"context"
	"fmt"
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
// cleans all extension logs older than the retention period when enabled
type ExtensionLogsCleanupJob struct {
	db     *bun.DB
	logger logger.Logger
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
// Since extension logs are system-wide (not org-scoped), this job cleans
// all extension logs older than the retention period specified in the org settings
func (j *ExtensionLogsCleanupJob) Run(ctx context.Context, orgID uuid.UUID) error {
	// Get organization settings to determine retention period
	var settings types.OrganizationSettings
	err := j.db.NewSelect().
		Model(&settings).
		Where("organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization settings: %w", err)
	}

	retentionDays := j.GetRetentionDays(&settings.Settings)
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Running extension logs cleanup (triggered by org %s, retention: %d days, cutoff: %s)",
			orgID, retentionDays, cutoffDate.Format(time.RFC3339)),
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
		fmt.Sprintf("Deleted %d extension logs (system-wide cleanup triggered by org %s)", rowsAffected, orgID),
		"",
	)

	return nil
}
