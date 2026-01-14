package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type HealthCheckStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

type HealthCheckRepository interface {
	CreateHealthCheck(healthCheck *shared_types.HealthCheck) error
	GetHealthCheckByApplicationID(applicationID uuid.UUID, organizationID uuid.UUID) (*shared_types.HealthCheck, error)
	GetHealthCheckByID(id uuid.UUID, organizationID uuid.UUID) (*shared_types.HealthCheck, error)
	UpdateHealthCheck(healthCheck *shared_types.HealthCheck) error
	DeleteHealthCheck(applicationID uuid.UUID, organizationID uuid.UUID) error
	ToggleHealthCheck(applicationID uuid.UUID, organizationID uuid.UUID, enabled bool) error
	GetEnabledHealthChecks() ([]*shared_types.HealthCheck, error)
	GetDueHealthChecks() ([]*shared_types.HealthCheck, error)
	AddHealthCheckResult(result *shared_types.HealthCheckResult) error
	GetHealthCheckResults(healthCheckID uuid.UUID, limit int, startTime, endTime *time.Time) ([]*shared_types.HealthCheckResult, error)
	GetHealthCheckStats(healthCheckID uuid.UUID, startTime, endTime time.Time) (*HealthCheckStats, error)
	CleanupOldResults(retentionDays int) error
	UpdateHealthCheckStatus(healthCheckID uuid.UUID, consecutiveFails int, lastCheckedAt time.Time) error
}

type HealthCheckStats struct {
	TotalChecks      int     `json:"total_checks"`
	SuccessfulChecks int     `json:"successful_checks"`
	FailedChecks     int     `json:"failed_checks"`
	UptimePercentage float64 `json:"uptime_percentage"`
	AvgResponseTime  int     `json:"avg_response_time_ms"`
}

func (s *HealthCheckStorage) CreateHealthCheck(healthCheck *shared_types.HealthCheck) error {
	_, err := s.DB.NewInsert().Model(healthCheck).Exec(s.Ctx)
	return err
}

func (s *HealthCheckStorage) GetHealthCheckByApplicationID(applicationID uuid.UUID, organizationID uuid.UUID) (*shared_types.HealthCheck, error) {
	var healthCheck shared_types.HealthCheck
	err := s.DB.NewSelect().
		Model(&healthCheck).
		Where("application_id = ? AND organization_id = ?", applicationID, organizationID).
		Scan(s.Ctx)

	if err != nil {
		return nil, err
	}
	return &healthCheck, nil
}

func (s *HealthCheckStorage) GetHealthCheckByID(id uuid.UUID, organizationID uuid.UUID) (*shared_types.HealthCheck, error) {
	var healthCheck shared_types.HealthCheck
	err := s.DB.NewSelect().
		Model(&healthCheck).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Scan(s.Ctx)

	if err != nil {
		return nil, err
	}
	return &healthCheck, nil
}

func (s *HealthCheckStorage) UpdateHealthCheck(healthCheck *shared_types.HealthCheck) error {
	_, err := s.DB.NewUpdate().
		Model(healthCheck).
		OmitZero().
		Set("updated_at = CURRENT_TIMESTAMP").
		WherePK().
		Exec(s.Ctx)
	return err
}

func (s *HealthCheckStorage) DeleteHealthCheck(applicationID uuid.UUID, organizationID uuid.UUID) error {
	_, err := s.DB.NewDelete().
		Model((*shared_types.HealthCheck)(nil)).
		Where("application_id = ? AND organization_id = ?", applicationID, organizationID).
		Exec(s.Ctx)
	return err
}

func (s *HealthCheckStorage) ToggleHealthCheck(applicationID uuid.UUID, organizationID uuid.UUID, enabled bool) error {
	_, err := s.DB.NewUpdate().
		Model((*shared_types.HealthCheck)(nil)).
		Set("enabled = ?", enabled).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("application_id = ? AND organization_id = ?", applicationID, organizationID).
		Exec(s.Ctx)
	return err
}

func (s *HealthCheckStorage) GetEnabledHealthChecks() ([]*shared_types.HealthCheck, error) {
	var healthChecks []*shared_types.HealthCheck
	err := s.DB.NewSelect().
		Model(&healthChecks).
		Where("enabled = ?", true).
		Scan(s.Ctx)
	return healthChecks, err
}

func (s *HealthCheckStorage) GetDueHealthChecks() ([]*shared_types.HealthCheck, error) {
	var healthChecks []*shared_types.HealthCheck
	now := time.Now()

	err := s.DB.NewSelect().
		Model(&healthChecks).
		Where("enabled = ?", true).
		Scan(s.Ctx)

	if err != nil {
		return nil, err
	}

	var dueChecks []*shared_types.HealthCheck
	for _, hc := range healthChecks {
		if hc.LastCheckedAt == nil {
			dueChecks = append(dueChecks, hc)
			continue
		}

		nextCheck := hc.LastCheckedAt.Add(time.Duration(hc.IntervalSeconds) * time.Second)
		if now.After(nextCheck) || now.Equal(nextCheck) {
			dueChecks = append(dueChecks, hc)
		}
	}

	return dueChecks, nil
}

func (s *HealthCheckStorage) AddHealthCheckResult(result *shared_types.HealthCheckResult) error {
	_, err := s.DB.NewInsert().Model(result).Exec(s.Ctx)
	return err
}

func (s *HealthCheckStorage) GetHealthCheckResults(healthCheckID uuid.UUID, limit int, startTime, endTime *time.Time) ([]*shared_types.HealthCheckResult, error) {
	var results []*shared_types.HealthCheckResult
	query := s.DB.NewSelect().
		Model(&results).
		Where("health_check_id = ?", healthCheckID).
		Order("checked_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if startTime != nil {
		query = query.Where("checked_at >= ?", *startTime)
	}

	if endTime != nil {
		query = query.Where("checked_at <= ?", *endTime)
	}

	err := query.Scan(s.Ctx)
	return results, err
}

func (s *HealthCheckStorage) GetHealthCheckStats(healthCheckID uuid.UUID, startTime, endTime time.Time) (*HealthCheckStats, error) {
	var stats HealthCheckStats

	// Get total checks count
	totalCount, err := s.DB.NewSelect().
		Model((*shared_types.HealthCheckResult)(nil)).
		Where("health_check_id = ?", healthCheckID).
		Where("checked_at >= ?", startTime).
		Where("checked_at <= ?", endTime).
		Count(s.Ctx)

	if err != nil {
		return nil, err
	}

	stats.TotalChecks = totalCount

	if stats.TotalChecks == 0 {
		return &stats, nil
	}

	// Get successful checks (healthy status)
	successfulCount, err := s.DB.NewSelect().
		Model((*shared_types.HealthCheckResult)(nil)).
		Where("health_check_id = ?", healthCheckID).
		Where("checked_at >= ?", startTime).
		Where("checked_at <= ?", endTime).
		Where("status = ?", "healthy").
		Count(s.Ctx)

	if err != nil {
		return nil, err
	}

	stats.SuccessfulChecks = successfulCount

	stats.FailedChecks = stats.TotalChecks - stats.SuccessfulChecks

	if stats.TotalChecks > 0 {
		stats.UptimePercentage = float64(stats.SuccessfulChecks) / float64(stats.TotalChecks) * 100
	}

	// Get average response time
	var avgResponseTime float64
	err = s.DB.NewSelect().
		Model((*shared_types.HealthCheckResult)(nil)).
		ColumnExpr("AVG(response_time_ms)").
		Where("health_check_id = ?", healthCheckID).
		Where("checked_at >= ?", startTime).
		Where("checked_at <= ?", endTime).
		Where("response_time_ms IS NOT NULL").
		Scan(s.Ctx, &avgResponseTime)

	if err == nil {
		stats.AvgResponseTime = int(avgResponseTime)
	}

	return &stats, nil
}

func (s *HealthCheckStorage) CleanupOldResults(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	_, err := s.DB.NewDelete().
		Model((*shared_types.HealthCheckResult)(nil)).
		Where("checked_at < ?", cutoffTime).
		Exec(s.Ctx)

	return err
}

func (s *HealthCheckStorage) UpdateHealthCheckStatus(healthCheckID uuid.UUID, consecutiveFails int, lastCheckedAt time.Time) error {
	_, err := s.DB.NewUpdate().
		Model((*shared_types.HealthCheck)(nil)).
		Set("consecutive_fails = ?", consecutiveFails).
		Set("last_checked_at = ?", lastCheckedAt).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("id = ?", healthCheckID).
		Exec(s.Ctx)
	return err
}
