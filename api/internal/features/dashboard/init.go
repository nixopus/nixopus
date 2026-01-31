package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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

func NewDashboardMonitor(conn *websocket.Conn, log logger.Logger, organizationID string, deployService DeployServiceProvider) (*DashboardMonitor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Get organization-specific SSH client
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, orgID.String())
	manager, err := sshpkg.GetSSHManagerFromContext(orgCtx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get SSH manager: %w", err)
	}

	dockerService, err := getDockerService()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get docker service: %w", err)
	}

	monitor := &DashboardMonitor{
		conn:           conn,
		sshManager:     manager,
		log:            log,
		ctx:            ctx,
		cancel:         cancel,
		Interval:       time.Second * 10,
		Operations:     AllOperations,
		dockerService:  dockerService,
		organizationID: organizationID,
		deployService:  deployService,
	}

	return monitor, nil
}

func (m *DashboardMonitor) Start() {
	go func() {
		ticker := time.NewTicker(m.Interval)
		defer ticker.Stop()

		// Get connection from pool (will reuse existing or create new)
		client, err := m.sshManager.Connect()
		if err != nil {
			m.log.Log(logger.Error, "Failed to connect to SSH server", err.Error())
			m.BroadcastError(err.Error(), "ssh_connect")
			return
		}
		m.client = client
		m.HandleAllOperations()

		for {
			select {
			case <-ticker.C:
				// Reuse existing connection from pool, reconnect if needed
				client, err := m.sshManager.Connect()
				if err != nil {
					m.log.Log(logger.Error, "Failed to reconnect to SSH server", err.Error())
					m.BroadcastError(err.Error(), "ssh_connect")
					continue
				}
				m.client = client
				m.HandleAllOperations()
			case <-m.ctx.Done():
				m.log.Log(logger.Info, "Dashboard monitor stopped", "")
				return
			}
		}
	}()
}

func (m *DashboardMonitor) Stop() {
	m.log.Log(logger.Info, "Stopping dashboard monitor", "")
	if m.cancel != nil {
		m.cancel()
	}
	m.Close()
}

func (m *DashboardMonitor) HandleAllOperations() {
	// Check if operations are already running - skip if so to prevent concurrent execution
	m.operationsMutex.Lock()
	if m.operationsRunning {
		m.operationsMutex.Unlock()
		return
	}
	m.operationsRunning = true
	m.operationsMutex.Unlock()

	// Ensure we reset the flag when done
	defer func() {
		m.operationsMutex.Lock()
		m.operationsRunning = false
		m.operationsMutex.Unlock()
	}()

	for _, operation := range m.Operations {
		select {
		case <-m.ctx.Done():
			return
		default:
			m.HandleOperation(operation)
		}
	}
}

func (m *DashboardMonitor) HandleOperation(operation DashboardOperation) {
	switch operation {
	case GetContainers:
		m.GetContainers()
	case GetSystemStats:
		m.GetSystemStats()
	case GetDeployments:
		m.GetDeployments()
	default:
		m.log.Log(logger.Error, "Unknown operation", string(operation))
	}
}

func (m *DashboardMonitor) UpdateConfig(config MonitoringConfig) {
	m.Interval = config.Interval

	if len(config.Operations) > 0 {
		m.Operations = config.Operations
	}

	m.Stop()
	m.Start()
}

func (m *DashboardMonitor) SetOperations(operations []DashboardOperation) {
	m.Operations = operations

	m.Stop()
	m.Start()
}

func (m *DashboardMonitor) Close() {
	// Don't close SSH client here - it's managed by the connection pool
	// The pool will handle cleanup when connections are idle
	m.client = nil

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
