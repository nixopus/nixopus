package caddy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
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

// ---- GetServerHealth and GetAllHealth with non-empty state ----------------

func TestHealthMonitor_GetServerHealth_knownOrg(t *testing.T) {
	l := logger.NewLogger()
	rec := NewReconciler(&deployRepoTestStub{}, l)
	h := NewHealthMonitor(l, rec, time.Minute, func(context.Context) ([]uuid.UUID, error) {
		return nil, nil
	})

	orgID := uuid.New()
	h.servers[orgID] = &ServerHealth{OrganizationID: orgID, Host: "h1", Healthy: true}

	s := h.GetServerHealth(orgID)
	require.NotNil(t, s)
	assert.Equal(t, "h1", s.Host)
	assert.True(t, s.Healthy)
}

func TestHealthMonitor_GetAllHealth_withEntries(t *testing.T) {
	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	id1, id2 := uuid.New(), uuid.New()
	h.servers[id1] = &ServerHealth{OrganizationID: id1, Host: "s1"}
	h.servers[id2] = &ServerHealth{OrganizationID: id2, Host: "s2"}

	all := h.GetAllHealth()
	assert.Len(t, all, 2)
}

// ---- updateHealth ---------------------------------------------------------

func TestHealthMonitor_updateHealth_newOrg_unhealthy(t *testing.T) {
	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	orgID := uuid.New()
	h.updateHealth(orgID, "srv1", false, "ping failed")

	s := h.GetServerHealth(orgID)
	require.NotNil(t, s)
	assert.False(t, s.Healthy)
	assert.Equal(t, 1, s.FailCount)
	assert.Equal(t, "ping failed", s.LastError)
}

func TestHealthMonitor_updateHealth_existingOrg_healthy(t *testing.T) {
	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	orgID := uuid.New()
	// Pre-populate with an unhealthy state.
	h.servers[orgID] = &ServerHealth{
		OrganizationID:  orgID,
		Healthy:         false,
		FailCount:       5,
		RestartAttempts: 2,
	}

	h.updateHealth(orgID, "srv1", true, "")

	s := h.GetServerHealth(orgID)
	require.NotNil(t, s)
	assert.True(t, s.Healthy)
	assert.Equal(t, 0, s.FailCount)
	assert.Equal(t, 0, s.RestartAttempts)
}

// ---- shouldAttemptRestart -------------------------------------------------

func TestHealthMonitor_shouldAttemptRestart(t *testing.T) {
	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	orgID := uuid.New()

	// No entry → true.
	assert.True(t, h.shouldAttemptRestart(orgID))

	// Entry with RestartAttempts=0 → true.
	h.servers[orgID] = &ServerHealth{RestartAttempts: 0}
	assert.True(t, h.shouldAttemptRestart(orgID))

	// Recent restart attempt → false.
	h.servers[orgID] = &ServerHealth{
		RestartAttempts:    1,
		LastRestartAttempt: time.Now(),
	}
	assert.False(t, h.shouldAttemptRestart(orgID))

	// Old restart attempt (past backoff) → true.
	h.servers[orgID] = &ServerHealth{
		RestartAttempts:    1,
		LastRestartAttempt: time.Now().Add(-5 * time.Minute),
	}
	assert.True(t, h.shouldAttemptRestart(orgID))
}

// ---- recordRestartAttempt -------------------------------------------------

func TestHealthMonitor_recordRestartAttempt(t *testing.T) {
	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	orgID := uuid.New()
	// No state → no-op.
	h.recordRestartAttempt(orgID, false)

	// With state.
	h.servers[orgID] = &ServerHealth{OrganizationID: orgID, RestartAttempts: 0}
	h.recordRestartAttempt(orgID, true)

	s := h.GetServerHealth(orgID)
	require.NotNil(t, s)
	assert.Equal(t, 1, s.RestartAttempts)
	assert.True(t, s.ContainerMissing)
	assert.False(t, s.LastRestartAttempt.IsZero())
}

// ---- triggerReconciliation ------------------------------------------------

func TestHealthMonitor_triggerReconciliation(t *testing.T) {
	origEnq := enqueueReconcileAfterRecovery
	t.Cleanup(func() { enqueueReconcileAfterRecovery = origEnq })

	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	// Hook returns error → log warning path.
	enqueueReconcileAfterRecovery = func(uuid.UUID) error { return errors.New("queue err") }
	h.triggerReconciliation(uuid.New())

	// Hook returns nil → success path.
	enqueueReconcileAfterRecovery = func(uuid.UUID) error { return nil }
	h.triggerReconciliation(uuid.New())
}

// ---- healthCheckTask ------------------------------------------------------

func TestHealthMonitor_healthCheckTask_invalidOrgID(t *testing.T) {
	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	err := h.healthCheckTask(context.Background(), CaddyTaskPayload{OrganizationID: "not-a-uuid"})
	require.Error(t, err)
}

func TestHealthMonitor_healthCheckTask_validOrgID(t *testing.T) {
	origHook := getSSHHostForOrgHook
	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "", errors.New("no ssh")
	}
	t.Cleanup(func() { getSSHHostForOrgHook = origHook })

	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	err := h.healthCheckTask(context.Background(), CaddyTaskPayload{OrganizationID: uuid.New().String()})
	require.NoError(t, err)
}

// ---- enqueueAll fallback (queue not initialized) --------------------------

func TestHealthMonitor_enqueueAll_fallbackDirectCheck(t *testing.T) {
	// Ensure health queue is not initialized so fallback path is taken.
	origQ := HealthCheckQueue
	origT := TaskCaddyHealthCheck
	HealthCheckQueue = nil
	TaskCaddyHealthCheck = nil
	t.Cleanup(func() {
		HealthCheckQueue = origQ
		TaskCaddyHealthCheck = origT
	})

	origHook := getSSHHostForOrgHook
	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "", errors.New("no ssh host")
	}
	t.Cleanup(func() { getSSHHostForOrgHook = origHook })

	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) {
			return []uuid.UUID{uuid.New()}, nil
		},
	)
	// Should not panic; fallback calls checkServer which fails early at getSSHHostForOrg.
	h.enqueueAll(context.Background())
}

// ---- checkServer paths via direct call ------------------------------------

func TestHealthMonitor_checkServer_sshHostError(t *testing.T) {
	origHook := getSSHHostForOrgHook
	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "", errors.New("no ssh")
	}
	t.Cleanup(func() { getSSHHostForOrgHook = origHook })

	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	orgID := uuid.New()
	h.checkServer(context.Background(), orgID)

	s := h.GetServerHealth(orgID)
	require.NotNil(t, s)
	assert.False(t, s.Healthy)
}

func TestHealthMonitor_checkServer_pingFailure_lowFailCount(t *testing.T) {
	origHook := getSSHHostForOrgHook
	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "srv.example", nil
	}
	t.Cleanup(func() { getSSHHostForOrgHook = origHook })

	origPing := pingCaddyProbe
	pingCaddyProbe = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) error {
		return errors.New("ping failed")
	}
	t.Cleanup(func() { pingCaddyProbe = origPing })

	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	orgID := uuid.New()
	h.checkServer(context.Background(), orgID)

	s := h.GetServerHealth(orgID)
	require.NotNil(t, s)
	assert.False(t, s.Healthy)
	assert.Equal(t, 1, s.FailCount)
}

func TestHealthMonitor_checkServer_pingSuccess_firstCheck(t *testing.T) {
	origHook := getSSHHostForOrgHook
	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "srv.example", nil
	}
	t.Cleanup(func() { getSSHHostForOrgHook = origHook })

	origPing := pingCaddyProbe
	pingCaddyProbe = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) error {
		return nil
	}
	t.Cleanup(func() { pingCaddyProbe = origPing })

	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	orgID := uuid.New()
	h.checkServer(context.Background(), orgID)

	s := h.GetServerHealth(orgID)
	require.NotNil(t, s)
	assert.True(t, s.Healthy)
}

func TestHealthMonitor_checkServer_pingSuccess_recovery(t *testing.T) {
	origHook := getSSHHostForOrgHook
	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "srv.example", nil
	}
	t.Cleanup(func() { getSSHHostForOrgHook = origHook })

	origPing := pingCaddyProbe
	pingCaddyProbe = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) error {
		return nil
	}
	t.Cleanup(func() { pingCaddyProbe = origPing })

	origEnq := enqueueReconcileAfterRecovery
	enqueueReconcileAfterRecovery = func(uuid.UUID) error { return nil }
	t.Cleanup(func() { enqueueReconcileAfterRecovery = origEnq })

	l := logger.NewLogger()
	h := NewHealthMonitor(l, NewReconciler(&deployRepoTestStub{}, l), time.Minute,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })

	orgID := uuid.New()
	// Pre-populate with an unhealthy state so "recovery" path fires (!wasHealthy && existed).
	h.servers[orgID] = &ServerHealth{OrganizationID: orgID, Healthy: false}

	h.checkServer(context.Background(), orgID)

	s := h.GetServerHealth(orgID)
	require.NotNil(t, s)
	assert.True(t, s.Healthy)
}
