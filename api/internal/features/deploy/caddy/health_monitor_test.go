package caddy

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthMonitor_GetServerHealth_unknownOrg(t *testing.T) {
	l := logger.NewLogger()
	rec := NewReconciler(&deployRepoTestStub{}, l)
	h := NewHealthMonitor(l, rec, time.Minute, func(context.Context) ([]uuid.UUID, error) {
		return nil, nil
	})

	assert.Nil(t, h.GetServerHealth(uuid.New()))
	assert.Empty(t, h.GetAllHealth())
}

func TestHealthMonitor_SetupQueue_idempotent(t *testing.T) {
	orig := TaskCaddyHealthCheck
	t.Cleanup(func() { TaskCaddyHealthCheck = orig })
	TaskCaddyHealthCheck = nil

	l := logger.NewLogger()
	rec := NewReconciler(&deployRepoTestStub{}, l)
	h := NewHealthMonitor(l, rec, time.Hour, func(context.Context) ([]uuid.UUID, error) {
		return nil, nil
	})

	h.SetupQueue()
	first := TaskCaddyHealthCheck
	require.NotNil(t, first)

	h.SetupQueue()
	assert.Same(t, first, TaskCaddyHealthCheck)
}
