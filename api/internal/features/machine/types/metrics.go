package types

import "time"

// MachineMetricRow is one raw metric sample.
type MachineMetricRow struct {
	Time         time.Time `json:"time"`
	MachineName  string    `json:"machine_name"`
	CPUUsagePct  *float32  `json:"cpu_usage_pct,omitempty"`
	MemUsagePct  *float32  `json:"mem_usage_pct,omitempty"`
	MemUsedMB    *float32  `json:"mem_used_mb,omitempty"`
	NetRxBytes   *int64    `json:"net_rx_bytes,omitempty"`
	NetTxBytes   *int64    `json:"net_tx_bytes,omitempty"`
	NetDrops     *int64    `json:"net_drops,omitempty"`
	BwDroppedPct *float32  `json:"bw_dropped_pct,omitempty"`
	ActiveConns  *int32    `json:"active_conns,omitempty"`
}

// MachineEventRow is one event record.
type MachineEventRow struct {
	Time        time.Time `json:"time"`
	MachineName string    `json:"machine_name"`
	EventType   string    `json:"event_type"`
	Detail      *string   `json:"detail,omitempty"`
}

// MachineTrafficHitRow is one traffic hit bucket.
type MachineTrafficHitRow struct {
	BucketStart  time.Time `json:"bucket_start"`
	MachineName  string    `json:"machine_name"`
	ReqCount     int64     `json:"req_count"`
	ErrCount     int64     `json:"err_count"`
	AvgLatencyMS *float64  `json:"avg_latency_ms,omitempty"`
}

// MachineSummary is an LLM-friendly aggregate summary of machine health.
type MachineSummary struct {
	MachineName   string            `json:"machine_name"`
	WindowStart   time.Time         `json:"window_start"`
	WindowEnd     time.Time         `json:"window_end"`
	AvgCPUPct     *float64          `json:"avg_cpu_pct,omitempty"`
	MaxCPUPct     *float64          `json:"max_cpu_pct,omitempty"`
	AvgMemPct     *float64          `json:"avg_mem_pct,omitempty"`
	MaxMemPct     *float64          `json:"max_mem_pct,omitempty"`
	TotalRxMB     *float64          `json:"total_rx_mb,omitempty"`
	TotalTxMB     *float64          `json:"total_tx_mb,omitempty"`
	TotalNetDrops *int64            `json:"total_net_drops,omitempty"`
	EventCount    int               `json:"event_count"`
	TotalReqCount int64             `json:"total_req_count"`
	TotalErrCount int64             `json:"total_err_count"`
	RecentEvents  []MachineEventRow `json:"recent_events"`
}

// MachineMetricsResponse wraps a list of metric rows.
type MachineMetricsResponse struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Data    []MachineMetricRow `json:"data"`
}

// MachineEventsResponse wraps a list of event rows.
type MachineEventsResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    []MachineEventRow `json:"data"`
}

// MachineSummaryResponse wraps the summary.
type MachineSummaryResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    MachineSummary `json:"data"`
}
