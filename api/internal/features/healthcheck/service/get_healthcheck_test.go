package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHealthCheck_InvalidApplicationID(t *testing.T) {
	svc := newTestService(&mockHealthCheckRepo{})
	_, err := svc.GetHealthCheck("bad-uuid", uuid.New())
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestGetHealthCheck_NotFound_ReturnsNilNil(t *testing.T) {
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return nil, sql.ErrNoRows
		},
	}
	svc := newTestService(repo)
	hc, err := svc.GetHealthCheck(uuid.New().String(), uuid.New())
	require.NoError(t, err)
	assert.Nil(t, hc)
}

func TestGetHealthCheck_OtherStorageError(t *testing.T) {
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return nil, errors.New("connection refused")
		},
	}
	svc := newTestService(repo)
	_, err := svc.GetHealthCheck(uuid.New().String(), uuid.New())
	assert.ErrorIs(t, err, types.ErrHealthCheckNotFound)
}

func TestGetHealthCheck_Success(t *testing.T) {
	appID := uuid.New()
	orgID := uuid.New()
	expected := &shared_types.HealthCheck{ID: uuid.New(), ApplicationID: appID}
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			assert.Equal(t, appID, a)
			assert.Equal(t, orgID, o)
			return expected, nil
		},
	}
	svc := newTestService(repo)
	hc, err := svc.GetHealthCheck(appID.String(), orgID)
	require.NoError(t, err)
	assert.Equal(t, expected, hc)
}
