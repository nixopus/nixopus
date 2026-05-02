package service

import (
	"errors"
	"testing"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDueHealthChecks_Success(t *testing.T) {
	checks := []*shared_types.HealthCheck{
		{Enabled: true},
		{Enabled: true},
	}
	repo := &mockHealthCheckRepo{
		getDueHealthChecks: func() ([]*shared_types.HealthCheck, error) {
			return checks, nil
		},
	}
	svc := newTestService(repo)
	got, err := svc.GetDueHealthChecks()
	require.NoError(t, err)
	assert.Equal(t, checks, got)
}

func TestGetDueHealthChecks_StorageError(t *testing.T) {
	storageErr := errors.New("db error")
	repo := &mockHealthCheckRepo{
		getDueHealthChecks: func() ([]*shared_types.HealthCheck, error) {
			return nil, storageErr
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetDueHealthChecks()
	assert.ErrorIs(t, err, storageErr)
}
