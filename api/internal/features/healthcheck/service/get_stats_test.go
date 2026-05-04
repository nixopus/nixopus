package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	hcstorage "github.com/nixopus/nixopus/api/internal/features/healthcheck/storage"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHealthCheckStats_InvalidApplicationID(t *testing.T) {
	svc := newTestService(&mockHealthCheckRepo{})
	_, err := svc.GetHealthCheckStats("bad-uuid", uuid.New(), "24h")
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestGetHealthCheckStats_HealthCheckNotFound(t *testing.T) {
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetHealthCheckStats(uuid.New().String(), uuid.New(), "24h")
	assert.ErrorIs(t, err, types.ErrHealthCheckNotFound)
}

func TestGetHealthCheckStats_Period_1h(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "1h", "1h")
}

func TestGetHealthCheckStats_Period_1H(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "1H", "1H")
}

func TestGetHealthCheckStats_Period_24h(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "24h", "24h")
}

func TestGetHealthCheckStats_Period_24H(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "24H", "24H")
}

func TestGetHealthCheckStats_Period_1d(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "1d", "1d")
}

func TestGetHealthCheckStats_Period_1D(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "1D", "1D")
}

func TestGetHealthCheckStats_Period_7d(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "7d", "7d")
}

func TestGetHealthCheckStats_Period_7D(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "7D", "7D")
}

func TestGetHealthCheckStats_Period_30d(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "30d", "30d")
}

func TestGetHealthCheckStats_Period_30D(t *testing.T) {
	testGetHealthCheckStatsPeriod(t, "30D", "30D")
}

func TestGetHealthCheckStats_Period_Default(t *testing.T) {
	hcID := uuid.New()
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: hcID}, nil
		},
		getHealthCheckStats: func(id uuid.UUID, st, et time.Time) (*hcstorage.HealthCheckStats, error) {
			return &hcstorage.HealthCheckStats{TotalChecks: 10, SuccessfulChecks: 8}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			return nil, nil
		},
	}
	svc := newTestService(repo)
	resp, err := svc.GetHealthCheckStats(uuid.New().String(), uuid.New(), "unknown-period")
	require.NoError(t, err)
	// Unrecognized period normalizes to "24h"
	assert.Equal(t, "24h", resp.Period)
}

func TestGetHealthCheckStats_StatsStorageError(t *testing.T) {
	statsErr := errors.New("stats query failed")
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		getHealthCheckStats: func(id uuid.UUID, st, et time.Time) (*hcstorage.HealthCheckStats, error) {
			return nil, statsErr
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetHealthCheckStats(uuid.New().String(), uuid.New(), "1h")
	assert.ErrorIs(t, err, statsErr)
}

func TestGetHealthCheckStats_LastStatusFromResults(t *testing.T) {
	hcID := uuid.New()
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: hcID}, nil
		},
		getHealthCheckStats: func(id uuid.UUID, st, et time.Time) (*hcstorage.HealthCheckStats, error) {
			return &hcstorage.HealthCheckStats{TotalChecks: 5, SuccessfulChecks: 5, UptimePercentage: 100}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			return []*shared_types.HealthCheckResult{
				{Status: string(shared_types.HealthCheckStatusHealthy)},
			}, nil
		},
	}
	svc := newTestService(repo)
	resp, err := svc.GetHealthCheckStats(uuid.New().String(), uuid.New(), "7d")
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp.LastStatus)
	assert.Equal(t, float64(100), resp.UptimePercentage)
}

func TestGetHealthCheckStats_LastStatusUnknownWhenNoResults(t *testing.T) {
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		getHealthCheckStats: func(id uuid.UUID, st, et time.Time) (*hcstorage.HealthCheckStats, error) {
			return &hcstorage.HealthCheckStats{}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			return []*shared_types.HealthCheckResult{}, nil
		},
	}
	svc := newTestService(repo)
	resp, err := svc.GetHealthCheckStats(uuid.New().String(), uuid.New(), "30d")
	require.NoError(t, err)
	assert.Equal(t, string(shared_types.HealthCheckStatusUnknown), resp.LastStatus)
}

func TestGetHealthCheckStats_LastStatusUnknownOnResultsError(t *testing.T) {
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		getHealthCheckStats: func(id uuid.UUID, st, et time.Time) (*hcstorage.HealthCheckStats, error) {
			return &hcstorage.HealthCheckStats{}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			return nil, errors.New("results query failed")
		},
	}
	svc := newTestService(repo)
	resp, err := svc.GetHealthCheckStats(uuid.New().String(), uuid.New(), "1h")
	require.NoError(t, err)
	assert.Equal(t, string(shared_types.HealthCheckStatusUnknown), resp.LastStatus)
}

// testGetHealthCheckStatsPeriod is a shared helper that verifies the period
// string is preserved in the response for recognised period values.
func testGetHealthCheckStatsPeriod(t *testing.T, period, expectedPeriod string) {
	t.Helper()
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		getHealthCheckStats: func(id uuid.UUID, st, et time.Time) (*hcstorage.HealthCheckStats, error) {
			return &hcstorage.HealthCheckStats{}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			return nil, nil
		},
	}
	svc := newTestService(repo)
	resp, err := svc.GetHealthCheckStats(uuid.New().String(), uuid.New(), period)
	require.NoError(t, err)
	assert.Equal(t, expectedPeriod, resp.Period)
}
