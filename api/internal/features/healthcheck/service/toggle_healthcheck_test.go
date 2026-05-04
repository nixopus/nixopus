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

func TestToggleHealthCheck_InvalidApplicationID(t *testing.T) {
	svc := newTestService(&mockHealthCheckRepo{})
	_, err := svc.ToggleHealthCheck(uuid.New(), &types.ToggleHealthCheckRequest{
		ApplicationID: "bad-uuid",
		Enabled:       true,
	})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestToggleHealthCheck_ToggleStorageError(t *testing.T) {
	toggleErr := errors.New("toggle failed")
	repo := &mockHealthCheckRepo{
		toggleHealthCheck: func(appID, orgID uuid.UUID, enabled bool) error {
			return toggleErr
		},
	}
	svc := newTestService(repo)
	_, err := svc.ToggleHealthCheck(uuid.New(), &types.ToggleHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Enabled:       false,
	})
	assert.ErrorIs(t, err, toggleErr)
}

func TestToggleHealthCheck_GetAfterToggleError(t *testing.T) {
	repo := &mockHealthCheckRepo{
		toggleHealthCheck: func(appID, orgID uuid.UUID, enabled bool) error {
			return nil
		},
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return nil, errors.New("not found after toggle")
		},
	}
	svc := newTestService(repo)
	_, err := svc.ToggleHealthCheck(uuid.New(), &types.ToggleHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Enabled:       true,
	})
	assert.ErrorIs(t, err, types.ErrHealthCheckNotFound)
}

func TestToggleHealthCheck_Success(t *testing.T) {
	appID := uuid.New()
	orgID := uuid.New()
	expected := &shared_types.HealthCheck{ID: uuid.New(), Enabled: true}
	repo := &mockHealthCheckRepo{
		toggleHealthCheck: func(a, o uuid.UUID, enabled bool) error {
			assert.Equal(t, appID, a)
			assert.Equal(t, orgID, o)
			assert.True(t, enabled)
			return nil
		},
		getHealthCheckByApplicationID: func(a, o uuid.UUID) (*shared_types.HealthCheck, error) {
			return expected, nil
		},
	}
	svc := newTestService(repo)
	hc, err := svc.ToggleHealthCheck(orgID, &types.ToggleHealthCheckRequest{
		ApplicationID: appID.String(),
		Enabled:       true,
	})
	require.NoError(t, err)
	assert.Equal(t, expected, hc)
}
