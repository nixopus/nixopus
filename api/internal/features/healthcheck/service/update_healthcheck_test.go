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

func TestUpdateHealthCheck_InvalidApplicationID(t *testing.T) {
	svc := newTestService(&mockHealthCheckRepo{})
	_, err := svc.UpdateHealthCheck(uuid.New(), &types.UpdateHealthCheckRequest{
		ApplicationID: "not-a-uuid",
	})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestUpdateHealthCheck_HealthCheckNotFound(t *testing.T) {
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newTestService(repo)
	_, err := svc.UpdateHealthCheck(uuid.New(), &types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.ErrorIs(t, err, types.ErrHealthCheckNotFound)
}

func TestUpdateHealthCheck_NoFieldsChanged(t *testing.T) {
	appID := uuid.New()
	orgID := uuid.New()
	original := &shared_types.HealthCheck{
		ID:               uuid.New(),
		Endpoint:         "/old",
		Method:           "GET",
		ExpectedStatus:   []int{200},
		TimeoutSeconds:   30,
		IntervalSeconds:  60,
		FailureThreshold: 3,
		SuccessThreshold: 1,
		RetentionDays:    30,
	}
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return original, nil
		},
	}
	svc := newTestService(repo)
	hc, err := svc.UpdateHealthCheck(orgID, &types.UpdateHealthCheckRequest{
		ApplicationID: appID.String(),
	})
	require.NoError(t, err)
	// Fields should remain unchanged when request has zero values
	assert.Equal(t, "/old", hc.Endpoint)
	assert.Equal(t, "GET", hc.Method)
}

func TestUpdateHealthCheck_AllFieldsUpdated(t *testing.T) {
	appID := uuid.New()
	orgID := uuid.New()
	original := &shared_types.HealthCheck{
		ID: uuid.New(),
	}
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return original, nil
		},
	}
	svc := newTestService(repo)
	req := &types.UpdateHealthCheckRequest{
		ApplicationID:    appID.String(),
		Endpoint:         "/new-health",
		Method:           "POST",
		ExpectedStatus:   []int{200, 204},
		TimeoutSeconds:   15,
		IntervalSeconds:  90,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Headers:          map[string]string{"X-Key": "val"},
		Body:             `{"ok":true}`,
		RetentionDays:    14,
	}
	hc, err := svc.UpdateHealthCheck(orgID, req)
	require.NoError(t, err)
	assert.Equal(t, "/new-health", hc.Endpoint)
	assert.Equal(t, "POST", hc.Method)
	assert.Equal(t, []int{200, 204}, hc.ExpectedStatus)
	assert.Equal(t, 15, hc.TimeoutSeconds)
	assert.Equal(t, 90, hc.IntervalSeconds)
	assert.Equal(t, 5, hc.FailureThreshold)
	assert.Equal(t, 2, hc.SuccessThreshold)
	assert.Equal(t, map[string]string{"X-Key": "val"}, hc.Headers)
	assert.Equal(t, `{"ok":true}`, hc.Body)
	assert.Equal(t, 14, hc.RetentionDays)
}

func TestUpdateHealthCheck_StorageUpdateError(t *testing.T) {
	appID := uuid.New()
	updateErr := errors.New("db write failed")
	repo := &mockHealthCheckRepo{
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return &shared_types.HealthCheck{ID: uuid.New()}, nil
		},
		updateHealthCheck: func(hc *shared_types.HealthCheck) error {
			return updateErr
		},
	}
	svc := newTestService(repo)
	_, err := svc.UpdateHealthCheck(uuid.New(), &types.UpdateHealthCheckRequest{
		ApplicationID: appID.String(),
	})
	assert.ErrorIs(t, err, updateErr)
}
