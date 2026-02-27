package dashboard

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type DashboardOperation string

const (
	GetContainers  DashboardOperation = "get_containers"
	GetSystemStats DashboardOperation = "get_system_stats"
	GetDeployments DashboardOperation = "get_deployments"
)

var AllOperations = []DashboardOperation{
	GetContainers,
	GetSystemStats,
	GetDeployments,
}

const defaultPollerInterval = 10 * time.Second

type MonitoringConfig struct {
	Interval   time.Duration        `json:"interval"`
	Operations []DashboardOperation `json:"operations"`
}

// DeployServiceProvider defines the interface for fetching latest deployments.
type DeployServiceProvider interface {
	GetLatestDeployments(organizationID string, limit int) ([]shared_types.ApplicationDeployment, error)
}

// OrgPoller runs a single polling loop per organization. All DashboardMonitors
// for the same org share one poller, so only ONE SSH session is created per
// polling interval regardless of how many users have the dashboard open.
type OrgPoller struct {
	subMu          sync.RWMutex
	subscribers    map[*DashboardMonitor]struct{}
	sshManager     *sshpkg.SSHManager
	dockerService  docker.DockerRepository
	deployService  DeployServiceProvider
	organizationID string
	log            logger.Logger

	runMu   sync.Mutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

var (
	orgPollers   = make(map[string]*OrgPoller)
	orgPollersMu sync.Mutex
)

// DashboardMonitor is a thin wrapper around a WebSocket connection that
// subscribes to a shared OrgPoller. It does NOT run its own polling loop.
type DashboardMonitor struct {
	conn      *websocket.Conn
	connMutex *sync.Mutex // shared per-connection write mutex
	log       logger.Logger
	poller    *OrgPoller

	subMu      sync.Mutex
	subscribed bool

	// Kept for API compatibility with the realtime package.
	Interval   time.Duration
	Operations []DashboardOperation
}

type SystemStats struct {
	OSType        string       `json:"os_type"`
	Hostname      string       `json:"hostname"`
	CPUInfo       string       `json:"cpu_info"`
	CPUCores      int          `json:"cpu_cores"`
	CPU           CPUStats     `json:"cpu"`
	Memory        MemoryStats  `json:"memory"`
	Load          LoadStats    `json:"load"`
	Disk          DiskStats    `json:"disk"`
	Network       NetworkStats `json:"network"`
	KernelVersion string       `json:"kernel_version"`
	Architecture  string       `json:"architecture"`
	Timestamp     time.Time    `json:"timestamp"`
}

type CPUCore struct {
	CoreID int     `json:"core_id"`
	Usage  float64 `json:"usage"`
}

type CPUStats struct {
	Overall float64   `json:"overall"`
	PerCore []CPUCore `json:"per_core"`
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

type NetworkInterface struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytesSent"`
	BytesRecv   uint64 `json:"bytesRecv"`
	PacketsSent uint64 `json:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv"`
	ErrorIn     uint64 `json:"errorIn"`
	ErrorOut    uint64 `json:"errorOut"`
	DropIn      uint64 `json:"dropIn"`
	DropOut     uint64 `json:"dropOut"`
}

type NetworkStats struct {
	TotalBytesSent   uint64             `json:"totalBytesSent"`
	TotalBytesRecv   uint64             `json:"totalBytesRecv"`
	TotalPacketsSent uint64             `json:"totalPacketsSent"`
	TotalPacketsRecv uint64             `json:"totalPacketsRecv"`
	Interfaces       []NetworkInterface `json:"interfaces"`
	UploadSpeed      float64            `json:"uploadSpeed"`
	DownloadSpeed    float64            `json:"downloadSpeed"`
}
