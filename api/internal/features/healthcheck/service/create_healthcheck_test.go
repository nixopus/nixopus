package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateHealthCheck_InvalidApplicationID(t *testing.T) {
	svc := newTestService(&mockHealthCheckRepo{})
	_, err := svc.CreateHealthCheck(uuid.New(), uuid.New(), &types.CreateHealthCheckRequest{
		ApplicationID: "not-a-uuid",
	})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestCreateHealthCheck_AlreadyExists(t *testing.T) {
	appID := uuid.New()
	orgID := uuid.New()
	existing := &shared_types.HealthCheck{ID: uuid.New()}
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return existing, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.CreateHealthCheck(uuid.New(), orgID, &types.CreateHealthCheckRequest{
		ApplicationID: appID.String(),
	})
	assert.ErrorIs(t, err, types.ErrHealthCheckAlreadyExists)
}

func TestCreateHealthCheck_Defaults(t *testing.T) {
	appID := uuid.New()
	orgID := uuid.New()
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newTestService(repo)
	hc, err := svc.CreateHealthCheck(uuid.New(), orgID, &types.CreateHealthCheckRequest{
		ApplicationID: appID.String(),
	})
	require.NoError(t, err)
	assert.Equal(t, "/", hc.Endpoint)
	assert.Equal(t, "GET", hc.Method)
	assert.Equal(t, []int{200}, hc.ExpectedStatus)
	assert.Equal(t, 30, hc.TimeoutSeconds)
	assert.Equal(t, 60, hc.IntervalSeconds)
	assert.Equal(t, 3, hc.FailureThreshold)
	assert.Equal(t, 1, hc.SuccessThreshold)
	assert.Equal(t, 30, hc.RetentionDays)
	assert.True(t, hc.Enabled)
	assert.Equal(t, orgID, hc.OrganizationID)
}

func TestCreateHealthCheck_ExplicitValues(t *testing.T) {
	appID := uuid.New()
	orgID := uuid.New()
	headers := map[string]string{"Authorization": "Bearer token"}
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newTestService(repo)
	req := &types.CreateHealthCheckRequest{
		ApplicationID:    appID.String(),
		Endpoint:         "/health",
		Method:           "POST",
		ExpectedStatus:   []int{200, 201},
		TimeoutSeconds:   10,
		IntervalSeconds:  120,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Headers:          headers,
		Body:             `{"ping":true}`,
		RetentionDays:    7,
	}
	hc, err := svc.CreateHealthCheck(uuid.New(), orgID, req)
	require.NoError(t, err)
	assert.Equal(t, "/health", hc.Endpoint)
	assert.Equal(t, "POST", hc.Method)
	assert.Equal(t, []int{200, 201}, hc.ExpectedStatus)
	assert.Equal(t, 10, hc.TimeoutSeconds)
	assert.Equal(t, 120, hc.IntervalSeconds)
	assert.Equal(t, 5, hc.FailureThreshold)
	assert.Equal(t, 2, hc.SuccessThreshold)
	assert.Equal(t, headers, hc.Headers)
	assert.Equal(t, `{"ping":true}`, hc.Body)
	assert.Equal(t, 7, hc.RetentionDays)
}

func TestCreateHealthCheck_StorageError(t *testing.T) {
	appID := uuid.New()
	orgID := uuid.New()
	storageErr := errors.New("db write failed")
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return nil, errors.New("not found")
		},
		createHealthCheck: func(hc *shared_types.HealthCheck) error {
			return storageErr
		},
	}
	svc := newTestService(repo)
	_, err := svc.CreateHealthCheck(uuid.New(), orgID, &types.CreateHealthCheckRequest{
		ApplicationID: appID.String(),
	})
	assert.ErrorIs(t, err, storageErr)
}
