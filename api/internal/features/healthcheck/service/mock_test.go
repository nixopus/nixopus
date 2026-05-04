package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	hcstorage "github.com/nixopus/nixopus/api/internal/features/healthcheck/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// ---------------------------------------------------------------------------
// mockHealthCheckRepo — test double for storage.HealthCheckRepository
// ---------------------------------------------------------------------------

type mockHealthCheckRepo struct {
	createHealthCheck             func(hc *shared_types.HealthCheck) error
	getHealthCheckByApplicationID func(appID, orgID uuid.UUID) (*shared_types.HealthCheck, error)
	getHealthCheckByID            func(id, orgID uuid.UUID) (*shared_types.HealthCheck, error)
	updateHealthCheck             func(hc *shared_types.HealthCheck) error
	deleteHealthCheck             func(appID, orgID uuid.UUID) error
	toggleHealthCheck             func(appID, orgID uuid.UUID, enabled bool) error
	getEnabledHealthChecks        func() ([]*shared_types.HealthCheck, error)
	getDueHealthChecks            func() ([]*shared_types.HealthCheck, error)
	addHealthCheckResult          func(result *shared_types.HealthCheckResult) error
	getHealthCheckResults         func(healthCheckID uuid.UUID, limit int, startTime, endTime *time.Time) ([]*shared_types.HealthCheckResult, error)
	getHealthCheckStats           func(healthCheckID uuid.UUID, startTime, endTime time.Time) (*hcstorage.HealthCheckStats, error)
	cleanupOldResults             func(retentionDays int) error
	updateHealthCheckStatus       func(healthCheckID uuid.UUID, consecutiveFails int, lastCheckedAt time.Time) error
}

func (m *mockHealthCheckRepo) CreateHealthCheck(hc *shared_types.HealthCheck) error {
	if m.createHealthCheck != nil {
		return m.createHealthCheck(hc)
	}
	return nil
}

func (m *mockHealthCheckRepo) GetHealthCheckByApplicationID(appID, orgID uuid.UUID) (*shared_types.HealthCheck, error) {
	if m.getHealthCheckByApplicationID != nil {
		return m.getHealthCheckByApplicationID(appID, orgID)
	}
	return nil, nil
}

func (m *mockHealthCheckRepo) GetHealthCheckByID(id, orgID uuid.UUID) (*shared_types.HealthCheck, error) {
	if m.getHealthCheckByID != nil {
		return m.getHealthCheckByID(id, orgID)
	}
	return nil, nil
}

func (m *mockHealthCheckRepo) UpdateHealthCheck(hc *shared_types.HealthCheck) error {
	if m.updateHealthCheck != nil {
		return m.updateHealthCheck(hc)
	}
	return nil
}

func (m *mockHealthCheckRepo) DeleteHealthCheck(appID, orgID uuid.UUID) error {
	if m.deleteHealthCheck != nil {
		return m.deleteHealthCheck(appID, orgID)
	}
	return nil
}

func (m *mockHealthCheckRepo) ToggleHealthCheck(appID, orgID uuid.UUID, enabled bool) error {
	if m.toggleHealthCheck != nil {
		return m.toggleHealthCheck(appID, orgID, enabled)
	}
	return nil
}

func (m *mockHealthCheckRepo) GetEnabledHealthChecks() ([]*shared_types.HealthCheck, error) {
	if m.getEnabledHealthChecks != nil {
		return m.getEnabledHealthChecks()
	}
	return nil, nil
}

func (m *mockHealthCheckRepo) GetDueHealthChecks() ([]*shared_types.HealthCheck, error) {
	if m.getDueHealthChecks != nil {
		return m.getDueHealthChecks()
	}
	return nil, nil
}

func (m *mockHealthCheckRepo) AddHealthCheckResult(result *shared_types.HealthCheckResult) error {
	if m.addHealthCheckResult != nil {
		return m.addHealthCheckResult(result)
	}
	return nil
}

func (m *mockHealthCheckRepo) GetHealthCheckResults(healthCheckID uuid.UUID, limit int, startTime, endTime *time.Time) ([]*shared_types.HealthCheckResult, error) {
	if m.getHealthCheckResults != nil {
		return m.getHealthCheckResults(healthCheckID, limit, startTime, endTime)
	}
	return nil, nil
}

func (m *mockHealthCheckRepo) GetHealthCheckStats(healthCheckID uuid.UUID, startTime, endTime time.Time) (*hcstorage.HealthCheckStats, error) {
	if m.getHealthCheckStats != nil {
		return m.getHealthCheckStats(healthCheckID, startTime, endTime)
	}
	return &hcstorage.HealthCheckStats{}, nil
}

func (m *mockHealthCheckRepo) CleanupOldResults(retentionDays int) error {
	if m.cleanupOldResults != nil {
		return m.cleanupOldResults(retentionDays)
	}
	return nil
}

func (m *mockHealthCheckRepo) UpdateHealthCheckStatus(healthCheckID uuid.UUID, consecutiveFails int, lastCheckedAt time.Time) error {
	if m.updateHealthCheckStatus != nil {
		return m.updateHealthCheckStatus(healthCheckID, consecutiveFails, lastCheckedAt)
	}
	return nil
}

// ---------------------------------------------------------------------------
// mockAppProvider — test double for ApplicationProvider
// ---------------------------------------------------------------------------

type mockAppProvider struct {
	getApplicationById    func(id string, organizationID uuid.UUID) (shared_types.Application, error)
	getApplicationDomains func(applicationID uuid.UUID) ([]shared_types.ApplicationDomain, error)
}

func (m *mockAppProvider) GetApplicationById(id string, organizationID uuid.UUID) (shared_types.Application, error) {
	if m.getApplicationById != nil {
		return m.getApplicationById(id, organizationID)
	}
	return shared_types.Application{}, nil
}

func (m *mockAppProvider) GetApplicationDomains(applicationID uuid.UUID) ([]shared_types.ApplicationDomain, error) {
	if m.getApplicationDomains != nil {
		return m.getApplicationDomains(applicationID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestService(repo *mockHealthCheckRepo) *HealthCheckService {
	return &HealthCheckService{
		storage: repo,
		logger:  logger.NewLogger(),
		ctx:     context.Background(),
	}
}

func newTestServiceWithApp(repo *mockHealthCheckRepo, appProvider ApplicationProvider) *HealthCheckService {
	svc := newTestService(repo)
	svc.appProvider = appProvider
	return svc
}
