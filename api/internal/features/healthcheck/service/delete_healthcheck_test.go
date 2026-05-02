package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteHealthCheck_InvalidApplicationID(t *testing.T) {
	svc := newTestService(&mockHealthCheckRepo{})
	err := svc.DeleteHealthCheck("bad-uuid", uuid.New())
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestDeleteHealthCheck_StorageError(t *testing.T) {
	storageErr := errors.New("delete failed")
	repo := &mockHealthCheckRepo{
		deleteHealthCheck: func(appID, orgID uuid.UUID) error {
			return storageErr
		},
	}
	svc := newTestService(repo)
	err := svc.DeleteHealthCheck(uuid.New().String(), uuid.New())
	assert.ErrorIs(t, err, storageErr)
}

func TestDeleteHealthCheck_Success(t *testing.T) {
	appID := uuid.New()
	orgID := uuid.New()
	var deletedAppID, deletedOrgID uuid.UUID
	repo := &mockHealthCheckRepo{
		deleteHealthCheck: func(a, o uuid.UUID) error {
			deletedAppID = a
			deletedOrgID = o
			return nil
		},
	}
	svc := newTestService(repo)
	err := svc.DeleteHealthCheck(appID.String(), orgID)
	require.NoError(t, err)
	assert.Equal(t, appID, deletedAppID)
	assert.Equal(t, orgID, deletedOrgID)
}
