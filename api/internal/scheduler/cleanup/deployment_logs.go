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
	DeploymentLogsCleanupJobName   = "deployment_logs_cleanup"
	DefaultDeploymentLogsRetention = 30
)

// DeploymentLogsCleanupJob handles cleanup of old deployment/application logs
type DeploymentLogsCleanupJob struct {
	db     *bun.DB
	logger logger.Logger
}

// NewDeploymentLogsCleanupJob creates a new deployment logs cleanup job
func NewDeploymentLogsCleanupJob(db *bun.DB, l logger.Logger) *DeploymentLogsCleanupJob {
	return &DeploymentLogsCleanupJob{
		db:     db,
		logger: l,
	}
}

// Name returns the job name
func (j *DeploymentLogsCleanupJob) Name() string {
	return DeploymentLogsCleanupJobName
}

// IsEnabled checks if cleanup is enabled for this organization
func (j *DeploymentLogsCleanupJob) IsEnabled(settings *types.OrganizationSettingsData) bool {
	return IsEnabledOrDefault(settings.DeploymentLogsCleanupEnabled)
}

// GetRetentionDays returns the retention period from settings
func (j *DeploymentLogsCleanupJob) GetRetentionDays(settings *types.OrganizationSettingsData) int {
	return GetRetentionDaysOrDefault(settings.DeploymentLogsRetentionDays, DefaultDeploymentLogsRetention)
}

// Run executes the cleanup job for a specific organization
func (j *DeploymentLogsCleanupJob) Run(ctx context.Context, orgID uuid.UUID) error {
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
		fmt.Sprintf("Running deployment logs cleanup for org %s (retention: %d days, cutoff: %s)",
			orgID, retentionDays, cutoffDate.Format(time.RFC3339)),
		"",
	)

	// Delete old application logs for applications belonging to this organization
	// application_logs -> applications -> organization_id
	result, err := j.db.NewDelete().
		TableExpr("application_logs").
		Where("application_id IN (?)",
			j.db.NewSelect().
				TableExpr("applications").
				Column("id").
				Where("organization_id = ?", orgID),
		).
		Where("created_at < ?", cutoffDate).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete deployment logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Deleted %d deployment logs for org %s", rowsAffected, orgID),
		"",
	)

	return nil
}
