package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/dashboard"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMachineRegStore implements MachineRegStore for unit tests.
type mockMachineRegStore struct {
	getMachineIsActiveFn       func(serverID uuid.UUID) (bool, error)
	setMachineActiveFn         func(serverID uuid.UUID, active bool) error
	getProvisionResourcesFn    func(serverID uuid.UUID) (int, int, int, error)
	updateProvisionResourcesFn func(serverID uuid.UUID, vcpu, memMB, diskGB int) error
}

func (m *mockMachineRegStore) GetMachineIsActive(serverID uuid.UUID) (bool, error) {
	if m.getMachineIsActiveFn != nil {
		return m.getMachineIsActiveFn(serverID)
	}
	return true, nil
}

func (m *mockMachineRegStore) SetMachineActive(serverID uuid.UUID, active bool) error {
	if m.setMachineActiveFn != nil {
		return m.setMachineActiveFn(serverID, active)
	}
	return nil
}

func (m *mockMachineRegStore) GetProvisionResources(serverID uuid.UUID) (int, int, int, error) {
	if m.getProvisionResourcesFn != nil {
		return m.getProvisionResourcesFn(serverID)
	}
	return 0, 0, 0, nil
}

func (m *mockMachineRegStore) UpdateProvisionResources(serverID uuid.UUID, vcpu, memMB, diskGB int) error {
	if m.updateProvisionResourcesFn != nil {
		return m.updateProvisionResourcesFn(serverID, vcpu, memMB, diskGB)
	}
	return nil
}

// mockSSHCommandRunner implements SSHCommandRunner for unit tests.
type mockSSHRunner struct {
	runCommandFn func(cmd string) (string, error)
}

func (m *mockSSHRunner) RunCommand(cmd string) (string, error) {
	if m.runCommandFn != nil {
		return m.runCommandFn(cmd)
	}
	return "", nil
}

// newMachineServiceForTest creates a MachineService with injected mocks.
func newMachineServiceForTest(reg MachineRegStore, runner SSHCommandRunner) *MachineService {
	svc := &MachineService{
		regStore: reg,
		logger:   logger.NewLogger(),
	}
	if runner != nil {
		svc.sshRunnerFn = func(_ context.Context) (SSHCommandRunner, error) {
			return runner, nil
		}
	}
	return svc
}

func setStatsCollector(svc *MachineService, fn func(runner SSHCommandRunner) (dashboard.SystemStats, error)) {
	svc.collectStatsFn = fn
}

// ---------- ParseUptimeToState ----------

func TestParseUptimeToState_Valid(t *testing.T) {
	state := ParseUptimeToState("1234.56 5678.90")
	require.NotNil(t, state)
	assert.True(t, state.Active)
	assert.Equal(t, "Running", state.State)
	assert.Equal(t, int64(1234), state.UptimeSec)
}

func TestParseUptimeToState_Empty(t *testing.T) {
	state := ParseUptimeToState("")
	require.NotNil(t, state)
	assert.True(t, state.Active)
	assert.Equal(t, int64(0), state.UptimeSec)
}

func TestParseUptimeToState_Whitespace(t *testing.T) {
	state := ParseUptimeToState("   3600.0   ")
	require.NotNil(t, state)
	assert.Equal(t, int64(3600), state.UptimeSec)
}

// ---------- NewMachineService ----------

func TestNewMachineService(t *testing.T) {
	svc := NewMachineService(nil, context.Background(), logger.NewLogger(), nil)
	assert.NotNil(t, svc)
}

// ---------- GetMachineStatus ----------

func TestMachineService_GetMachineStatus_SSHError(t *testing.T) {
	// no sshRunnerFn → falls back to real SSH which fails without a configured org
	svc := NewMachineService(nil, context.Background(), logger.NewLogger(), nil)
	_, err := svc.GetMachineStatus(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to server")
}

func TestMachineService_GetMachineStatus_PausedServer(t *testing.T) {
	serverID := uuid.New()
	reg := &mockMachineRegStore{
		getMachineIsActiveFn: func(_ uuid.UUID) (bool, error) { return false, nil },
	}
	svc := newMachineServiceForTest(reg, nil)
	// no sshRunnerFn set, but we never reach SSH because of the paused-server early return

	// However, we need a runner to not panic if we reach SSH. But with regStore returning
	// inactive=false, we return before SSH is called.
	ctx := context.WithValue(context.Background(), shared_types.ServerIDKey, serverID.String())

	// Since no SSH runner is injected and regStore returns inactive=false,
	// the function should return the Paused state immediately.
	// But sshRunnerFn is nil so it would try real SSH if not early-returned.
	// We set a failing runner to prove SSH is NOT called.
	svc.sshRunnerFn = func(_ context.Context) (SSHCommandRunner, error) {
		t.Fatal("SSH should not be called for paused server")
		return nil, nil
	}

	resp, err := svc.GetMachineStatus(ctx, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Paused", resp.Data.State)
	assert.False(t, resp.Data.Active)
}

func TestMachineService_GetMachineStatus_Stopped(t *testing.T) {
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) { return "", fmt.Errorf("connection refused") },
	}
	svc := newMachineServiceForTest(nil, runner)

	resp, err := svc.GetMachineStatus(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Stopped", resp.Data.State)
	assert.False(t, resp.Data.Active)
}

func TestMachineService_GetMachineStatus_Running_NilRegStore(t *testing.T) {
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) { return "7200.5 14400.0", nil },
	}
	// nil regStore → lazyFillSpecs returns immediately
	svc := newMachineServiceForTest(nil, runner)

	resp, err := svc.GetMachineStatus(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Running", resp.Data.State)
	assert.True(t, resp.Data.Active)
	assert.Equal(t, int64(7200), resp.Data.UptimeSec)
}

func TestMachineService_GetMachineStatus_Running_NoServerID(t *testing.T) {
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) { return "3600.0 7200.0", nil },
	}
	reg := &mockMachineRegStore{}
	svc := newMachineServiceForTest(reg, runner)

	// no ServerIDKey in context → lazyFillSpecs returns early at uuid.Parse
	resp, err := svc.GetMachineStatus(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Running", resp.Data.State)
}

func TestMachineService_GetMachineStatus_LazyFill_ProvisionError(t *testing.T) {
	serverID := uuid.New()
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) { return "3600.0", nil },
	}
	reg := &mockMachineRegStore{
		getMachineIsActiveFn:    func(_ uuid.UUID) (bool, error) { return true, nil },
		getProvisionResourcesFn: func(_ uuid.UUID) (int, int, int, error) { return 0, 0, 0, fmt.Errorf("db error") },
	}
	svc := newMachineServiceForTest(reg, runner)

	ctx := context.WithValue(context.Background(), shared_types.ServerIDKey, serverID.String())
	resp, err := svc.GetMachineStatus(ctx, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Running", resp.Data.State)
}

func TestMachineService_GetMachineStatus_LazyFill_AlreadyHasResources(t *testing.T) {
	serverID := uuid.New()
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) { return "3600.0", nil },
	}
	reg := &mockMachineRegStore{
		getMachineIsActiveFn:    func(_ uuid.UUID) (bool, error) { return true, nil },
		getProvisionResourcesFn: func(_ uuid.UUID) (int, int, int, error) { return 4, 4096, 50, nil },
	}
	svc := newMachineServiceForTest(reg, runner)

	ctx := context.WithValue(context.Background(), shared_types.ServerIDKey, serverID.String())
	resp, err := svc.GetMachineStatus(ctx, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Running", resp.Data.State)
}

func TestMachineService_GetMachineStatus_LazyFill_CollectStatsError(t *testing.T) {
	serverID := uuid.New()
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) {
			if cmd == "cat /proc/uptime" {
				return "3600.0", nil
			}
			return "", fmt.Errorf("command not found")
		},
	}
	reg := &mockMachineRegStore{
		getMachineIsActiveFn:    func(_ uuid.UUID) (bool, error) { return true, nil },
		getProvisionResourcesFn: func(_ uuid.UUID) (int, int, int, error) { return 0, 0, 0, nil },
	}
	svc := newMachineServiceForTest(reg, runner)

	ctx := context.WithValue(context.Background(), shared_types.ServerIDKey, serverID.String())
	resp, err := svc.GetMachineStatus(ctx, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Running", resp.Data.State)
}

func TestMachineService_GetMachineStatus_LazyFill_ZeroSpecs_NoUpdate(t *testing.T) {
	serverID := uuid.New()
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) { return "3600.0", nil },
	}
	reg := &mockMachineRegStore{
		getMachineIsActiveFn:    func(_ uuid.UUID) (bool, error) { return true, nil },
		getProvisionResourcesFn: func(_ uuid.UUID) (int, int, int, error) { return 0, 0, 0, nil },
		updateProvisionResourcesFn: func(_ uuid.UUID, _, _, _ int) error {
			t.Fatal("should not update when computed specs are zero")
			return nil
		},
	}
	svc := newMachineServiceForTest(reg, runner)
	setStatsCollector(svc, func(_ SSHCommandRunner) (dashboard.SystemStats, error) {
		return dashboard.SystemStats{}, nil
	})

	ctx := context.WithValue(context.Background(), shared_types.ServerIDKey, serverID.String())
	resp, err := svc.GetMachineStatus(ctx, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Running", resp.Data.State)
}

func TestMachineService_GetMachineStatus_LazyFill_UpdateError(t *testing.T) {
	serverID := uuid.New()
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) { return "3600.0", nil },
	}
	reg := &mockMachineRegStore{
		getMachineIsActiveFn:    func(_ uuid.UUID) (bool, error) { return true, nil },
		getProvisionResourcesFn: func(_ uuid.UUID) (int, int, int, error) { return 0, 0, 0, nil },
		updateProvisionResourcesFn: func(_ uuid.UUID, _, _, _ int) error {
			return fmt.Errorf("persist failed")
		},
	}
	svc := newMachineServiceForTest(reg, runner)
	setStatsCollector(svc, func(_ SSHCommandRunner) (dashboard.SystemStats, error) {
		return dashboard.SystemStats{
			CPUCores: 2,
			Memory:   dashboard.MemoryStats{Total: 8},
			Disk:     dashboard.DiskStats{Total: 120},
		}, nil
	})

	ctx := context.WithValue(context.Background(), shared_types.ServerIDKey, serverID.String())
	resp, err := svc.GetMachineStatus(ctx, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Running", resp.Data.State)
}

// ---------- GetSystemStats ----------

func TestMachineService_GetSystemStats_SSHError(t *testing.T) {
	svc := NewMachineService(nil, context.Background(), logger.NewLogger(), nil)
	_, err := svc.GetSystemStats(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to server")
}

func TestMachineService_GetSystemStats_CollectError(t *testing.T) {
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) { return "", fmt.Errorf("ssh error") },
	}
	svc := newMachineServiceForTest(nil, runner)

	_, err := svc.GetSystemStats(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect system stats")
}

func TestMachineService_GetSystemStats_Success(t *testing.T) {
	runner := &mockSSHRunner{}
	svc := newMachineServiceForTest(nil, runner)
	setStatsCollector(svc, func(_ SSHCommandRunner) (dashboard.SystemStats, error) {
		return dashboard.SystemStats{
			CPUCores: 4,
			Memory:   dashboard.MemoryStats{Total: 16},
		}, nil
	})

	resp, err := svc.GetSystemStats(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, 4, resp.Data.CPUCores)
}

// ---------- RestartMachine ----------

func TestMachineService_RestartMachine_SSHError(t *testing.T) {
	svc := NewMachineService(nil, context.Background(), logger.NewLogger(), nil)
	_, err := svc.RestartMachine(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to server")
}

func TestMachineService_RestartMachine_Success(t *testing.T) {
	runner := &mockSSHRunner{
		runCommandFn: func(cmd string) (string, error) { return "", nil },
	}
	svc := newMachineServiceForTest(nil, runner)

	resp, err := svc.RestartMachine(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

// ---------- ExecCommand ----------

func TestMachineService_ExecCommand_SSHError(t *testing.T) {
	svc := NewMachineService(nil, context.Background(), logger.NewLogger(), nil)
	_, err := svc.ExecCommand(context.Background(), uuid.New(), "echo hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to server")
}

func TestMachineService_ExecCommand_InjectedSuccess(t *testing.T) {
	svc := newMachineServiceForTest(nil, nil)
	svc.sshExecFn = func(_ context.Context, _ string) (string, string, int, error) {
		return "ok", "warn", 42, nil
	}
	resp, err := svc.ExecCommand(context.Background(), uuid.New(), "echo hello")
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "ok", resp.Data.Stdout)
	assert.Equal(t, "warn", resp.Data.Stderr)
	assert.Equal(t, 42, resp.Data.ExitCode)
}

func TestMachineService_ExecCommand_InjectedError(t *testing.T) {
	svc := newMachineServiceForTest(nil, nil)
	svc.sshExecFn = func(_ context.Context, _ string) (string, string, int, error) {
		return "", "", 0, fmt.Errorf("exec failed")
	}
	_, err := svc.ExecCommand(context.Background(), uuid.New(), "echo hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command execution failed")
}

func TestMachineService_ExecCommand_RunnerError(t *testing.T) {
	runner := &mockSSHRunner{
		runCommandFn: func(_ string) (string, error) { return "", fmt.Errorf("runner failed") },
	}
	svc := newMachineServiceForTest(nil, runner)
	resp, err := svc.ExecCommand(context.Background(), uuid.New(), "echo hello")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "command execution failed")
}

func TestMachineService_ExecCommand_RunnerSuccess(t *testing.T) {
	runner := &mockSSHRunner{
		runCommandFn: func(_ string) (string, error) { return "hello", nil },
	}
	svc := newMachineServiceForTest(nil, runner)
	resp, err := svc.ExecCommand(context.Background(), uuid.New(), "echo hello")
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "hello", resp.Data.Stdout)
	assert.Equal(t, "", resp.Data.Stderr)
	assert.Equal(t, 0, resp.Data.ExitCode)
}

// ---------- PauseMachine ----------

func TestMachineService_PauseMachine_Success(t *testing.T) {
	serverID := uuid.New()
	reg := &mockMachineRegStore{
		setMachineActiveFn: func(_ uuid.UUID, _ bool) error { return nil },
	}
	svc := newMachineServiceForTest(reg, nil)

	resp, err := svc.PauseMachine(context.Background(), uuid.New(), serverID)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Contains(t, resp.Message, "paused")
}

func TestMachineService_PauseMachine_Error(t *testing.T) {
	serverID := uuid.New()
	reg := &mockMachineRegStore{
		setMachineActiveFn: func(_ uuid.UUID, _ bool) error { return fmt.Errorf("db error") },
	}
	svc := newMachineServiceForTest(reg, nil)

	_, err := svc.PauseMachine(context.Background(), uuid.New(), serverID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pause machine")
}

// ---------- ResumeMachine ----------

func TestMachineService_ResumeMachine_Success(t *testing.T) {
	serverID := uuid.New()
	reg := &mockMachineRegStore{
		setMachineActiveFn: func(_ uuid.UUID, _ bool) error { return nil },
	}
	svc := newMachineServiceForTest(reg, nil)

	resp, err := svc.ResumeMachine(context.Background(), uuid.New(), serverID)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Contains(t, resp.Message, "resumed")
}

func TestMachineService_ResumeMachine_Error(t *testing.T) {
	serverID := uuid.New()
	reg := &mockMachineRegStore{
		setMachineActiveFn: func(_ uuid.UUID, _ bool) error { return fmt.Errorf("db error") },
	}
	svc := newMachineServiceForTest(reg, nil)

	_, err := svc.ResumeMachine(context.Background(), uuid.New(), serverID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resume machine")
}
