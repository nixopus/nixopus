package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/melbahja/goph"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"

	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

func getDockerService() (*docker.DockerService, error) {
	service, err := docker.GetDockerManager().GetDefaultService()
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, fmt.Errorf("docker service is nil")
	}
	return service, nil
}

type ApplicationMonitor struct {
	conn          *websocket.Conn
	connMutex     sync.Mutex
	sshpkg        *sshpkg.SSH
	log           logger.Logger
	client        *goph.Client
	Interval      time.Duration
	cancel        context.CancelFunc
	ctx           context.Context
	dockerService *docker.DockerService
	Operations    []ApplicationMonitorOperation
}

type MonitoringConfig struct {
	Interval   time.Duration                 `json:"interval"`
	Operations []ApplicationMonitorOperation `json:"operations"`
}

type ApplicationMonitorOperation string

const (
	ContainerStatistics ApplicationMonitorOperation = "container_statistics"
)

func NewApplicationMonitor(conn *websocket.Conn, log logger.Logger) (*ApplicationMonitor, error) {
	ssh_client := sshpkg.NewSSH()
	ctx, cancel := context.WithCancel(context.Background())

	dockerService, err := getDockerService()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get docker service: %w", err)
	}

	monitor := &ApplicationMonitor{
		conn:          conn,
		sshpkg:        ssh_client,
		log:           log,
		ctx:           ctx,
		cancel:        cancel,
		Interval:      time.Second * 10,
		dockerService: dockerService,
		Operations:    []ApplicationMonitorOperation{ContainerStatistics},
	}

	return monitor, nil
}

func (m *ApplicationMonitor) Start() {
	go func() {
		ticker := time.NewTicker(m.Interval)
		defer ticker.Stop()
		client, err := m.sshpkg.Connect()
		if err != nil {
			m.log.Log(logger.Error, "Failed to connect to SSH server", err.Error())
			m.BroadcastError(err.Error(), "ssh_connect")
			return
		}
		m.client = client
		defer client.Close()

		for {
			select {
			case <-ticker.C:
				m.HandleAllOperations()
			case <-m.ctx.Done():
				m.log.Log(logger.Info, "Dashboard monitor stopped", "")
				return
			}
		}
	}()
}

func (m *ApplicationMonitor) Broadcast(action string, message interface{}) {
	lockAcquired := make(chan bool, 1)
	go func() {
		m.connMutex.Lock()
		lockAcquired <- true
	}()

	select {
	case <-lockAcquired:
		defer m.connMutex.Unlock()
		if m.conn == nil {
			m.log.Log(logger.Error, "WebSocket connection is nil", "")
			return
		}
		_ = m.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		if err := m.conn.WriteJSON(map[string]interface{}{"action": action,
			"data":      message,
			"timestamp": time.Now().Unix(),
			"topic":     "monitor_application"}); err != nil {
			m.log.Log(logger.Error, "Failed to broadcast message", err.Error())
		}

		_ = m.conn.SetWriteDeadline(time.Time{})

	case <-time.After(3 * time.Second):
		m.log.Log(logger.Error, "Timeout waiting for broadcast lock", "")
	}
}

func (m *ApplicationMonitor) BroadcastDebug(message string) {
	response := map[string]interface{}{
		"action":    "debug",
		"message":   message,
		"timestamp": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		m.log.Log(logger.Error, "Failed to marshal debug message", err.Error())
		return
	}

	m.Broadcast("debug", string(jsonData))
}

func (m *ApplicationMonitor) BroadcastError(errMsg string, operation ApplicationMonitorOperation) {
	response := map[string]interface{}{
		"action":    operation,
		"error":     errMsg,
		"timestamp": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		m.log.Log(logger.Error, "Failed to marshal error message", err.Error())
		return
	}

	m.Broadcast("error", string(jsonData))
}

func (m *ApplicationMonitor) HandleAllOperations() {
	for _, operation := range m.Operations {
		select {
		case <-m.ctx.Done():
			return
		default:
			m.HandleOperation(operation)
		}
	}
}

func (m *ApplicationMonitor) HandleOperation(operation ApplicationMonitorOperation) {
	switch operation {
	case ContainerStatistics:
		m.GetContainerStatistics()
	default:
		m.log.Log(logger.Error, "Unknown operation", string(operation))
	}
}

func (m *ApplicationMonitor) GetContainerStatistics() {
	containers, err := m.dockerService.ListAllContainers()
	if err != nil {
		m.log.Log(logger.Error, "Failed to get containers", err.Error())
		m.BroadcastError(err.Error(), "container_statistics")
		return
	}
	m.Broadcast(string(ContainerStatistics), containers)
}

func (m *ApplicationMonitor) UpdateConfig(config MonitoringConfig) {
	m.Interval = config.Interval

	if len(config.Operations) > 0 {
		m.Operations = config.Operations
	}

	m.Stop()
	m.Start()
}

func (m *ApplicationMonitor) SetOperations(operations []ApplicationMonitorOperation) {
	m.Operations = operations

	m.Stop()
	m.Start()
}

func (m *ApplicationMonitor) Close() {
	if m.client != nil {
		m.client.Close()
		m.client = nil
	}

	if m.conn != nil {
		m.connMutex.Lock()
		_ = m.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing connection"),
		)
		_ = m.conn.Close()
		m.conn = nil
		m.connMutex.Unlock()
	}
}

func (m *ApplicationMonitor) Stop() {
	m.log.Log(logger.Info, "Stopping application monitor", "")
	if m.cancel != nil {
		m.cancel()
	}
	m.Close()
}
