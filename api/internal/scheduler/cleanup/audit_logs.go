package cleanup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

const (
	AuditLogsCleanupJobName   = "audit_logs_cleanup"
	DefaultAuditLogsRetention = 30
)

// AuditLogsCleanupJob handles cleanup of old audit logs
type AuditLogsCleanupJob struct {
	db     *bun.DB
	logger logger.Logger
}

// NewAuditLogsCleanupJob creates a new audit logs cleanup job
func NewAuditLogsCleanupJob(db *bun.DB, l logger.Logger) *AuditLogsCleanupJob {
	return &AuditLogsCleanupJob{
		db:     db,
		logger: l,
	}
}

// Name returns the job name
func (j *AuditLogsCleanupJob) Name() string {
	return AuditLogsCleanupJobName
}

// IsEnabled checks if cleanup is enabled for this organization
func (j *AuditLogsCleanupJob) IsEnabled(settings *types.OrganizationSettingsData) bool {
	return IsEnabledOrDefault(settings.AuditLogsCleanupEnabled)
}

// GetRetentionDays returns the retention period from settings
func (j *AuditLogsCleanupJob) GetRetentionDays(settings *types.OrganizationSettingsData) int {
	return GetRetentionDaysOrDefault(settings.AuditLogsRetentionDays, DefaultAuditLogsRetention)
}

// Run executes the cleanup job for a specific organization
func (j *AuditLogsCleanupJob) Run(ctx context.Context, orgID uuid.UUID) error {
	// Get organization settings to determine retention period
	var settings types.OrganizationSettings
	err := j.db.NewSelect().
		Model(&settings).
		Where("organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No settings found, use default retention period
			j.logger.Log(
				logger.Info,
				fmt.Sprintf("No settings found for org %s, using default retention", orgID),
				"",
			)
		} else {
			return fmt.Errorf("failed to get organization settings: %w", err)
		}
	}

	retentionDays := j.GetRetentionDays(&settings.Settings)
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Running audit logs cleanup for org %s (retention: %d days, cutoff: %s)",
			orgID, retentionDays, cutoffDate.Format(time.RFC3339)),
		"",
	)

	// Delete old audit logs for this organization
	// audit_logs has organization_id directly
	result, err := j.db.NewDelete().
		TableExpr("audit_logs").
		Where("organization_id = ?", orgID).
		Where("created_at < ?", cutoffDate).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete audit logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Deleted %d audit logs for org %s", rowsAffected, orgID),
		"",
	)

	return nil
}
