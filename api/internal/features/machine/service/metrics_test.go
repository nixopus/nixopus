package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/service"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- mock TimescaleQuerier ----------

type mockTimescaleQuerier struct {
	getMetricsFn func(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time, limit int) ([]machine_types.MachineMetricRow, error)
	getEventsFn  func(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time, limit int) ([]machine_types.MachineEventRow, error)
	getSummaryFn func(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time) (*machine_types.MachineSummary, error)
}

func (m *mockTimescaleQuerier) GetMetrics(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time, limit int) ([]machine_types.MachineMetricRow, error) {
	if m.getMetricsFn != nil {
		return m.getMetricsFn(ctx, machineName, orgID, from, to, limit)
	}
	return []machine_types.MachineMetricRow{}, nil
}

func (m *mockTimescaleQuerier) GetEvents(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time, limit int) ([]machine_types.MachineEventRow, error) {
	if m.getEventsFn != nil {
		return m.getEventsFn(ctx, machineName, orgID, from, to, limit)
	}
	return []machine_types.MachineEventRow{}, nil
}

func (m *mockTimescaleQuerier) GetSummary(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time) (*machine_types.MachineSummary, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, machineName, orgID, from, to)
	}
	return &machine_types.MachineSummary{}, nil
}

func newTestMetricsService(ts *mockTimescaleQuerier, resolverFn func(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (string, error)) *service.MetricsService {
	return service.NewMetricsServiceWith(ts, resolverFn)
}

// ---------- constructor ----------

func TestNewMetricsService(t *testing.T) {
	svc := service.NewMetricsService(nil, nil)
	assert.NotNil(t, svc)
}

func TestNewMetricsServiceWith(t *testing.T) {
	ts := &mockTimescaleQuerier{}
	svc := service.NewMetricsServiceWith(ts, nil)
	assert.NotNil(t, svc)
}

// ---------- resolveMachineName (via GetMetrics) ----------

func TestMetricsService_ResolveError_PropagatesFromGetMetrics(t *testing.T) {
	ts := &mockTimescaleQuerier{}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (string, error) {
		return "", fmt.Errorf("no machine found")
	})
	_, err := svc.GetMetrics(context.Background(), uuid.New(), nil, time.Now().Add(-1*time.Hour), time.Now(), 100)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no machine found")
}

// ---------- GetMetrics ----------

func TestMetricsService_GetMetrics_Success(t *testing.T) {
	rows := []machine_types.MachineMetricRow{
		{MachineName: "container-1"},
	}
	ts := &mockTimescaleQuerier{
		getMetricsFn: func(_ context.Context, _ string, _ uuid.UUID, _, _ time.Time, _ int) ([]machine_types.MachineMetricRow, error) {
			return rows, nil
		},
	}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (string, error) {
		return "container-1", nil
	})
	resp, err := svc.GetMetrics(context.Background(), uuid.New(), nil, time.Now().Add(-1*time.Hour), time.Now(), 100)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Len(t, resp.Data, 1)
}

func TestMetricsService_GetMetrics_StorageError(t *testing.T) {
	ts := &mockTimescaleQuerier{
		getMetricsFn: func(_ context.Context, _ string, _ uuid.UUID, _, _ time.Time, _ int) ([]machine_types.MachineMetricRow, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (string, error) {
		return "container-1", nil
	})
	_, err := svc.GetMetrics(context.Background(), uuid.New(), nil, time.Now().Add(-1*time.Hour), time.Now(), 100)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// ---------- GetEvents ----------

func TestMetricsService_GetEvents_ResolveError(t *testing.T) {
	ts := &mockTimescaleQuerier{}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (string, error) {
		return "", fmt.Errorf("resolve error")
	})
	_, err := svc.GetEvents(context.Background(), uuid.New(), nil, time.Now().Add(-1*time.Hour), time.Now(), 100)
	require.Error(t, err)
}

func TestMetricsService_GetEvents_Success(t *testing.T) {
	events := []machine_types.MachineEventRow{
		{MachineName: "container-1", EventType: "start"},
	}
	ts := &mockTimescaleQuerier{
		getEventsFn: func(_ context.Context, _ string, _ uuid.UUID, _, _ time.Time, _ int) ([]machine_types.MachineEventRow, error) {
			return events, nil
		},
	}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (string, error) {
		return "container-1", nil
	})
	resp, err := svc.GetEvents(context.Background(), uuid.New(), nil, time.Now().Add(-1*time.Hour), time.Now(), 100)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Len(t, resp.Data, 1)
}

func TestMetricsService_GetEvents_StorageError(t *testing.T) {
	ts := &mockTimescaleQuerier{
		getEventsFn: func(_ context.Context, _ string, _ uuid.UUID, _, _ time.Time, _ int) ([]machine_types.MachineEventRow, error) {
			return nil, fmt.Errorf("events db error")
		},
	}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (string, error) {
		return "container-1", nil
	})
	_, err := svc.GetEvents(context.Background(), uuid.New(), nil, time.Now().Add(-1*time.Hour), time.Now(), 100)
	require.Error(t, err)
}

// ---------- GetSummary ----------

func TestMetricsService_GetSummary_ResolveError(t *testing.T) {
	ts := &mockTimescaleQuerier{}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (string, error) {
		return "", fmt.Errorf("resolve error")
	})
	_, err := svc.GetSummary(context.Background(), uuid.New(), nil, time.Now().Add(-1*time.Hour), time.Now())
	require.Error(t, err)
}

func TestMetricsService_GetSummary_Success(t *testing.T) {
	cpu := 25.5
	summary := &machine_types.MachineSummary{AvgCPUPct: &cpu}
	ts := &mockTimescaleQuerier{
		getSummaryFn: func(_ context.Context, _ string, _ uuid.UUID, _, _ time.Time) (*machine_types.MachineSummary, error) {
			return summary, nil
		},
	}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (string, error) {
		return "container-1", nil
	})
	resp, err := svc.GetSummary(context.Background(), uuid.New(), nil, time.Now().Add(-1*time.Hour), time.Now())
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	if assert.NotNil(t, resp.Data.AvgCPUPct) {
		assert.InDelta(t, 25.5, *resp.Data.AvgCPUPct, 0.001)
	}
}

func TestMetricsService_GetSummary_StorageError(t *testing.T) {
	ts := &mockTimescaleQuerier{
		getSummaryFn: func(_ context.Context, _ string, _ uuid.UUID, _, _ time.Time) (*machine_types.MachineSummary, error) {
			return nil, fmt.Errorf("summary db error")
		},
	}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (string, error) {
		return "container-1", nil
	})
	_, err := svc.GetSummary(context.Background(), uuid.New(), nil, time.Now().Add(-1*time.Hour), time.Now())
	require.Error(t, err)
}

// ---------- resolveMachineName via serverID and orgID paths ----------

func TestMetricsService_GetMetrics_WithServerID(t *testing.T) {
	serverID := uuid.New()
	calledWithServerID := false
	ts := &mockTimescaleQuerier{}
	svc := newTestMetricsService(ts, func(_ context.Context, _ uuid.UUID, sid *uuid.UUID) (string, error) {
		if sid != nil && *sid == serverID {
			calledWithServerID = true
		}
		return "container-1", nil
	})
	resp, err := svc.GetMetrics(context.Background(), uuid.New(), &serverID, time.Now().Add(-1*time.Hour), time.Now(), 100)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.True(t, calledWithServerID)
}
