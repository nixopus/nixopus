package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/service"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvisionInfoProvider struct {
	info *storage.ProvisionInfo
	err  error
}

func (m *mockProvisionInfoProvider) GetProvisionInfo(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (*storage.ProvisionInfo, error) {
	return m.info, m.err
}

func mockRPC(result *queue.MachineLifecycleResult, err error) service.LifecycleExecutor {
	return func(ctx context.Context, payload queue.MachineLifecyclePayload) (*queue.MachineLifecycleResult, error) {
		return result, err
	}
}

func TestLifecycleService_GetStatus_NotProvisioned(t *testing.T) {
	svc := service.NewLifecycleService(
		&mockProvisionInfoProvider{info: nil, err: nil},
		mockRPC(nil, nil),
	)

	_, err := svc.GetStatus(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineNotProvisioned)
}

func TestLifecycleService_GetStatus_Success(t *testing.T) {
	statusData, _ := json.Marshal(map[string]interface{}{
		"name": "trail-abc", "active": true, "state": "Running", "pid": 1234, "uptime_sec": 3600,
	})

	svc := service.NewLifecycleService(
		&mockProvisionInfoProvider{
			info: &storage.ProvisionInfo{
				UserID:        uuid.New(),
				ContainerName: "trail-abc",
				ServerID:      "srv-1",
			},
		},
		mockRPC(&queue.MachineLifecycleResult{
			Success: true,
			Action:  "status",
			Data:    statusData,
		}, nil),
	)

	resp, err := svc.GetStatus(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.True(t, resp.Data.Active)
	assert.Equal(t, "Running", resp.Data.State)
	assert.Equal(t, 1234, resp.Data.PID)
}

func TestLifecycleService_GetStatus_WithServerID(t *testing.T) {
	statusData, _ := json.Marshal(map[string]interface{}{
		"name": "trail-abc", "active": true, "state": "Running", "pid": 5678, "uptime_sec": 7200,
	})

	serverID := uuid.New()
	svc := service.NewLifecycleService(
		&mockProvisionInfoProvider{
			info: &storage.ProvisionInfo{
				UserID:        uuid.New(),
				ContainerName: "trail-abc",
				ServerID:      serverID.String(),
			},
		},
		mockRPC(&queue.MachineLifecycleResult{
			Success: true,
			Action:  "status",
			Data:    statusData,
		}, nil),
	)

	resp, err := svc.GetStatus(context.Background(), uuid.New(), &serverID)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.True(t, resp.Data.Active)
	assert.Equal(t, 5678, resp.Data.PID)
}

func TestLifecycleService_Restart_Success(t *testing.T) {
	svc := service.NewLifecycleService(
		&mockProvisionInfoProvider{
			info: &storage.ProvisionInfo{ContainerName: "trail-xyz", ServerID: "srv-2"},
		},
		mockRPC(&queue.MachineLifecycleResult{Success: true, Action: "restart"}, nil),
	)

	resp, err := svc.Restart(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Contains(t, resp.Message, "restart")
}

func TestLifecycleService_Pause_Success(t *testing.T) {
	svc := service.NewLifecycleService(
		&mockProvisionInfoProvider{
			info: &storage.ProvisionInfo{ContainerName: "trail-xyz", ServerID: "srv-2"},
		},
		mockRPC(&queue.MachineLifecycleResult{Success: true, Action: "pause"}, nil),
	)

	resp, err := svc.Pause(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

func TestLifecycleService_Resume_Success(t *testing.T) {
	svc := service.NewLifecycleService(
		&mockProvisionInfoProvider{
			info: &storage.ProvisionInfo{ContainerName: "trail-xyz", ServerID: "srv-2"},
		},
		mockRPC(&queue.MachineLifecycleResult{Success: true, Action: "resume"}, nil),
	)

	resp, err := svc.Resume(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

func TestLifecycleService_RPCTimeout(t *testing.T) {
	svc := service.NewLifecycleService(
		&mockProvisionInfoProvider{
			info: &storage.ProvisionInfo{ContainerName: "trail-xyz", ServerID: "srv-2"},
		},
		mockRPC(nil, fmt.Errorf("machine operation timed out")),
	)

	_, err := svc.Restart(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineOperationTimeout)
}

func TestLifecycleService_RPCFailure(t *testing.T) {
	svc := service.NewLifecycleService(
		&mockProvisionInfoProvider{
			info: &storage.ProvisionInfo{ContainerName: "trail-xyz", ServerID: "srv-2"},
		},
		mockRPC(&queue.MachineLifecycleResult{
			Success: false,
			Action:  "pause",
			Error:   "another operation is in progress for this instance",
		}, nil),
	)

	_, err := svc.Pause(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineOperationLocked)
}

func TestLifecycleService_EmptyContainerName(t *testing.T) {
	svc := service.NewLifecycleService(
		&mockProvisionInfoProvider{
			info: &storage.ProvisionInfo{ContainerName: "", ServerID: "srv-1"},
		},
		mockRPC(nil, nil),
	)

	_, err := svc.GetStatus(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineNotProvisioned)
}
