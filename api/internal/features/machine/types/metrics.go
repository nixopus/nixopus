package types

import "time"

// MachineMetricRow is one raw metric sample from vm_metrics.
type MachineMetricRow struct {
	Time             time.Time `json:"time"`
	MachineName      string    `json:"machine_name"`
	CPUUsagePct      *float32  `json:"cpu_usage_pct,omitempty"`
	MemoryBytes      *int64    `json:"memory_bytes,omitempty"`
	MemoryLimitBytes *int64    `json:"memory_limit_bytes,omitempty"`
	NetRxBytes       *int64    `json:"net_rx_bytes,omitempty"`
	NetTxBytes       *int64    `json:"net_tx_bytes,omitempty"`
	NetRxPackets     *int64    `json:"net_rx_packets,omitempty"`
	NetTxPackets     *int64    `json:"net_tx_packets,omitempty"`
	NetRxDrops       *int64    `json:"net_rx_drops,omitempty"`
	NetTxDrops       *int64    `json:"net_tx_drops,omitempty"`
}

// MachineEventRow is one event record from vm_events.
type MachineEventRow struct {
	Time        time.Time `json:"time"`
	MachineName string    `json:"machine_name"`
	EventType   string    `json:"event_type"`
	Count       int       `json:"count"`
	Details     *string   `json:"details,omitempty"`
}

// MachineTrafficHitRow is one traffic hit record from vm_traffic_hits.
type MachineTrafficHitRow struct {
	Time         time.Time `json:"time"`
	MachineName  string    `json:"machine_name"`
	RequestCount int       `json:"request_count"`
	ErrorCount   int       `json:"error_count"`
	BytesSent    int64     `json:"bytes_sent"`
}

// MachineSummary is an LLM-friendly aggregate summary of machine health.
type MachineSummary struct {
	MachineName      string            `json:"machine_name"`
	WindowStart      time.Time         `json:"window_start"`
	WindowEnd        time.Time         `json:"window_end"`
	AvgCPUPct        *float64          `json:"avg_cpu_pct,omitempty"`
	MaxCPUPct        *float64          `json:"max_cpu_pct,omitempty"`
	AvgMemoryMB      *float64          `json:"avg_memory_mb,omitempty"`
	MaxMemoryMB      *float64          `json:"max_memory_mb,omitempty"`
	TotalRxMB        *float64          `json:"total_rx_mb,omitempty"`
	TotalTxMB        *float64          `json:"total_tx_mb,omitempty"`
	TotalRxDrops     *float64          `json:"total_rx_drops,omitempty"`
	TotalTxDrops     *float64          `json:"total_tx_drops,omitempty"`
	EventCount       int               `json:"event_count"`
	TotalReqCount    int64             `json:"total_req_count"`
	TotalErrCount    int64             `json:"total_err_count"`
	TotalBytesSentMB float64           `json:"total_bytes_sent_mb"`
	RecentEvents     []MachineEventRow `json:"recent_events"`
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
