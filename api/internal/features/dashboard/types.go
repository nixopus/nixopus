package dashboard

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/melbahja/goph"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

type DashboardOperation string

const (
	GetContainers  DashboardOperation = "get_containers"
	GetSystemStats DashboardOperation = "get_system_stats"
)

var AllOperations = []DashboardOperation{
	GetContainers,
	GetSystemStats,
}

type MonitoringConfig struct {
	Interval   time.Duration        `json:"interval"`
	Operations []DashboardOperation `json:"operations"`
}

type DashboardMonitor struct {
	conn          *websocket.Conn
	connMutex     sync.Mutex
	sshpkg        *sshpkg.SSH
	log           logger.Logger
	client        *goph.Client
	Interval      time.Duration
	Operations    []DashboardOperation
	cancel        context.CancelFunc
	ctx           context.Context
	dockerService *docker.DockerService
}

type SystemStats struct {
	OSType        string      `json:"os_type"`
	Hostname      string      `json:"hostname"`
	CPUInfo       string      `json:"cpu_info"`
	CPUCores      int         `json:"cpu_cores"`
	CPU           CPUStats    `json:"cpu"`
	Memory        MemoryStats `json:"memory"`
	Load          LoadStats   `json:"load"`
	Disk          DiskStats   `json:"disk"`
	KernelVersion string      `json:"kernel_version"`
	Architecture  string      `json:"architecture"`
	Timestamp     time.Time   `json:"timestamp"`
}

type CPUCore struct {
	CoreID     int     `json:"core_id"`
	Usage      float64 `json:"usage"`
}

type CPUStats struct {
	Overall     float64   `json:"overall"`
	PerCore     []CPUCore `json:"per_core"`
}

type MemoryStats struct {
	Used       float64 `json:"used"`
	Total      float64 `json:"total"`
	Percentage float64 `json:"percentage"`
	RawInfo    string  `json:"rawInfo"`
}

type LoadStats struct {
	OneMin     float64 `json:"oneMin"`
	FiveMin    float64 `json:"fiveMin"`
	FifteenMin float64 `json:"fifteenMin"`
	Uptime     string  `json:"uptime"`
}

type DiskMount struct {
	Filesystem string `json:"filesystem"`
	Size       string `json:"size"`
	Used       string `json:"used"`
	Avail      string `json:"avail"`
	Capacity   string `json:"capacity"`
	MountPoint string `json:"mountPoint"`
}

type DiskStats struct {
	Total      float64     `json:"total"`
	Used       float64     `json:"used"`
	Available  float64     `json:"available"`
	Percentage float64     `json:"percentage"`
	MountPoint string      `json:"mountPoint"`
	AllMounts  []DiskMount `json:"allMounts"`
}
