package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHealthCheckResults_InvalidApplicationID(t *testing.T) {
	svc := newTestService(&mockHealthCheckRepo{})
	_, err := svc.GetHealthCheckResults("bad-uuid", uuid.New(), 10, "", "")
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestGetHealthCheckResults_HealthCheckNotFound(t *testing.T) {
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetHealthCheckResults(uuid.New().String(), uuid.New(), 10, "", "")
	assert.ErrorIs(t, err, types.ErrHealthCheckNotFound)
}

func TestGetHealthCheckResults_DefaultLimit(t *testing.T) {
	var capturedLimit int
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			capturedLimit = limit
			return nil, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetHealthCheckResults(uuid.New().String(), uuid.New(), 0, "", "")
	require.NoError(t, err)
	assert.Equal(t, 100, capturedLimit)
}

func TestGetHealthCheckResults_NegativeLimitDefaultsTo100(t *testing.T) {
	var capturedLimit int
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			capturedLimit = limit
			return nil, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetHealthCheckResults(uuid.New().String(), uuid.New(), -5, "", "")
	require.NoError(t, err)
	assert.Equal(t, 100, capturedLimit)
}

func TestGetHealthCheckResults_ValidTimeRange(t *testing.T) {
	start := "2024-01-01T00:00:00Z"
	end := "2024-12-31T23:59:59Z"
	var capturedStart, capturedEnd *time.Time
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			capturedStart = st
			capturedEnd = et
			return nil, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetHealthCheckResults(uuid.New().String(), uuid.New(), 50, start, end)
	require.NoError(t, err)
	require.NotNil(t, capturedStart)
	require.NotNil(t, capturedEnd)
	assert.Equal(t, "2024-01-01 00:00:00 +0000 UTC", capturedStart.UTC().String())
	assert.Equal(t, "2024-12-31 23:59:59 +0000 UTC", capturedEnd.UTC().String())
}

func TestGetHealthCheckResults_InvalidTimeStringsIgnored(t *testing.T) {
	var capturedStart, capturedEnd *time.Time
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			capturedStart = st
			capturedEnd = et
			return nil, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetHealthCheckResults(uuid.New().String(), uuid.New(), 10, "not-a-date", "also-not-a-date")
	require.NoError(t, err)
	assert.Nil(t, capturedStart)
	assert.Nil(t, capturedEnd)
}

func TestGetHealthCheckResults_StorageError(t *testing.T) {
	storageErr := errors.New("db read failed")
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			return nil, storageErr
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetHealthCheckResults(uuid.New().String(), uuid.New(), 10, "", "")
	assert.ErrorIs(t, err, storageErr)
}

func TestGetHealthCheckResults_Success(t *testing.T) {
	hcID := uuid.New()
	results := []*shared_types.HealthCheckResult{
		{ID: uuid.New(), Status: "healthy"},
		{ID: uuid.New(), Status: "unhealthy"},
	}
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: hcID}, nil
		},
		getHealthCheckResults: func(id uuid.UUID, limit int, st, et *time.Time) ([]*shared_types.HealthCheckResult, error) {
			assert.Equal(t, hcID, id)
			return results, nil
		},
	}
	svc := newTestService(repo)
	got, err := svc.GetHealthCheckResults(uuid.New().String(), uuid.New(), 10, "", "")
	require.NoError(t, err)
	assert.Equal(t, results, got)
}
