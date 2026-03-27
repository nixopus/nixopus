package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
)

// TimescaleStore queries the Timescale metrics database with reader credentials.
type TimescaleStore struct {
	pool *pgxpool.Pool
}

// NewTimescaleStore opens a pgxpool connection to the given Timescale URL.
// Returns nil, nil when url is empty — all query methods gracefully return empty results.
func NewTimescaleStore(ctx context.Context, url string) (*TimescaleStore, error) {
	if url == "" {
		return nil, nil
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("timescale connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("timescale ping: %w", err)
	}
	return &TimescaleStore{pool: pool}, nil
}

// Close releases all pool connections.
func (ts *TimescaleStore) Close() {
	if ts != nil {
		ts.pool.Close()
	}
}

// GetMetrics queries vm_metrics (raw) or a continuous aggregate depending on window duration.
// machineName maps to the vm_name column in Timescale.
func (ts *TimescaleStore) GetMetrics(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time, limit int) ([]machine_types.MachineMetricRow, error) {
	if ts == nil {
		return []machine_types.MachineMetricRow{}, nil
	}

	// Use hourly aggregate for windows > 48h, minute aggregate for > 6h, raw otherwise.
	duration := to.Sub(from)
	var table string
	switch {
	case duration > 48*time.Hour:
		table = "vm_metrics_1hr"
	case duration > 6*time.Hour:
		table = "vm_metrics_1min"
	default:
		table = "vm_metrics"
	}

	if limit <= 0 || limit > 1000 {
		limit = 500
	}

	query := fmt.Sprintf(`
		SELECT time, vm_name, cpu_usage_pct, mem_usage_pct, mem_used_mb,
		       net_rx_bytes, net_tx_bytes, net_drops, bw_dropped_pct, active_conns
		FROM %s
		WHERE vm_name = $1 AND org_id = $2 AND time >= $3 AND time < $4
		ORDER BY time DESC
		LIMIT $5`, table)

	rows, err := ts.pool.Query(ctx, query, machineName, orgID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}
	defer rows.Close()

	var result []machine_types.MachineMetricRow
	for rows.Next() {
		var r machine_types.MachineMetricRow
		if err := rows.Scan(&r.Time, &r.MachineName, &r.CPUUsagePct, &r.MemUsagePct, &r.MemUsedMB,
			&r.NetRxBytes, &r.NetTxBytes, &r.NetDrops, &r.BwDroppedPct, &r.ActiveConns); err != nil {
			return nil, fmt.Errorf("scan metrics row: %w", err)
		}
		result = append(result, r)
	}
	if result == nil {
		result = []machine_types.MachineMetricRow{}
	}
	return result, rows.Err()
}

// GetEvents queries vm_events for the given machine and org.
func (ts *TimescaleStore) GetEvents(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time, limit int) ([]machine_types.MachineEventRow, error) {
	if ts == nil {
		return []machine_types.MachineEventRow{}, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	rows, err := ts.pool.Query(ctx, `
		SELECT time, vm_name, event_type, detail
		FROM vm_events
		WHERE vm_name = $1 AND org_id = $2 AND time >= $3 AND time < $4
		ORDER BY time DESC
		LIMIT $5`,
		machineName, orgID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var result []machine_types.MachineEventRow
	for rows.Next() {
		var r machine_types.MachineEventRow
		if err := rows.Scan(&r.Time, &r.MachineName, &r.EventType, &r.Detail); err != nil {
			return nil, fmt.Errorf("scan event row: %w", err)
		}
		result = append(result, r)
	}
	if result == nil {
		result = []machine_types.MachineEventRow{}
	}
	return result, rows.Err()
}

// GetSummary returns an aggregate summary of machine health over the window.
func (ts *TimescaleStore) GetSummary(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time) (*machine_types.MachineSummary, error) {
	if ts == nil {
		return &machine_types.MachineSummary{MachineName: machineName, WindowStart: from, WindowEnd: to}, nil
	}

	summary := &machine_types.MachineSummary{MachineName: machineName, WindowStart: from, WindowEnd: to}

	// Aggregate from minute-level rollup.
	if err := ts.pool.QueryRow(ctx, `
		SELECT AVG(cpu_usage_pct), MAX(cpu_usage_pct), AVG(mem_usage_pct), MAX(mem_usage_pct),
		       SUM(net_rx_bytes)/1048576.0, SUM(net_tx_bytes)/1048576.0, SUM(net_drops)
		FROM vm_metrics_1min
		WHERE vm_name = $1 AND org_id = $2 AND time >= $3 AND time < $4`,
		machineName, orgID, from, to,
	).Scan(&summary.AvgCPUPct, &summary.MaxCPUPct, &summary.AvgMemPct, &summary.MaxMemPct,
		&summary.TotalRxMB, &summary.TotalTxMB, &summary.TotalNetDrops); err != nil {
		return nil, fmt.Errorf("aggregate metrics: %w", err)
	}

	if err := ts.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM vm_events
		WHERE vm_name = $1 AND org_id = $2 AND time >= $3 AND time < $4`,
		machineName, orgID, from, to,
	).Scan(&summary.EventCount); err != nil {
		return nil, fmt.Errorf("count events: %w", err)
	}

	if err := ts.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(req_count),0), COALESCE(SUM(err_count),0)
		FROM vm_traffic_hits
		WHERE vm_name = $1 AND org_id = $2 AND bucket_start >= $3 AND bucket_start < $4`,
		machineName, orgID, from, to,
	).Scan(&summary.TotalReqCount, &summary.TotalErrCount); err != nil {
		return nil, fmt.Errorf("aggregate traffic: %w", err)
	}

	events, err := ts.GetEvents(ctx, machineName, orgID, from, to, 10)
	if err != nil {
		return nil, err
	}
	summary.RecentEvents = events
	return summary, nil
}
