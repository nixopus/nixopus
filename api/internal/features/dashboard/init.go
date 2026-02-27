package dashboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func getDockerService(ctx context.Context) (docker.DockerRepository, error) {
	service, err := docker.GetDockerServiceFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, fmt.Errorf("docker service is nil")
	}
	return service, nil
}

// getOrCreateOrgPoller returns the shared poller for an org, creating one if
// it doesn't already exist. The poller's polling loop is started lazily by
// subscribe when the first monitor joins.
func getOrCreateOrgPoller(
	organizationID string,
	sshManager *sshpkg.SSHManager,
	dockerService docker.DockerRepository,
	deployService DeployServiceProvider,
	log logger.Logger,
) *OrgPoller {
	orgPollersMu.Lock()
	defer orgPollersMu.Unlock()

	if poller, ok := orgPollers[organizationID]; ok {
		return poller
	}

	poller := &OrgPoller{
		subscribers:    make(map[*DashboardMonitor]struct{}),
		sshManager:     sshManager,
		dockerService:  dockerService,
		deployService:  deployService,
		organizationID: organizationID,
		log:            log,
	}
	orgPollers[organizationID] = poller
	return poller
}

// subscribe adds a monitor. If this is the first subscriber the polling loop
// is started automatically.
func (p *OrgPoller) subscribe(m *DashboardMonitor) {
	p.subMu.Lock()
	p.subscribers[m] = struct{}{}
	count := len(p.subscribers)
	p.subMu.Unlock()

	if count == 1 {
		p.start()
	}
}

// unsubscribe removes a monitor. When the last subscriber leaves the polling
// loop is stopped and the poller is removed from the global registry.
func (p *OrgPoller) unsubscribe(m *DashboardMonitor) {
	p.subMu.Lock()
	delete(p.subscribers, m)
	remaining := len(p.subscribers)
	p.subMu.Unlock()

	if remaining == 0 {
		p.stop()
	}
}

func (p *OrgPoller) start() {
	p.runMu.Lock()
	defer p.runMu.Unlock()
	if p.running {
		return
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.running = true
	go p.run()
}

func (p *OrgPoller) stop() {
	p.runMu.Lock()
	defer p.runMu.Unlock()
	if !p.running {
		return
	}
	p.cancel()
	p.running = false

	orgPollersMu.Lock()
	delete(orgPollers, p.organizationID)
	orgPollersMu.Unlock()
}

// run is the single polling goroutine for this organization.
func (p *OrgPoller) run() {
	ticker := time.NewTicker(defaultPollerInterval)
	defer ticker.Stop()

	p.collectAndBroadcast()

	for {
		select {
		case <-ticker.C:
			p.collectAndBroadcast()
		case <-p.ctx.Done():
			p.log.Log(logger.Info, "Org poller stopped", p.organizationID)
			return
		}
	}
}

func (p *OrgPoller) collectAndBroadcast() {
	for _, op := range AllOperations {
		select {
		case <-p.ctx.Done():
			return
		default:
			p.handleOperation(op)
		}
	}
}

func (p *OrgPoller) handleOperation(operation DashboardOperation) {
	switch operation {
	case GetContainers:
		p.getContainers()
	case GetSystemStats:
		p.getSystemStats()
	case GetDeployments:
		p.getDeployments()
	default:
		p.log.Log(logger.Error, "Unknown operation", string(operation))
	}
}

// NewDashboardMonitor creates a monitor that subscribes to the shared per-org
// poller. Call Start to begin receiving data and Stop to unsubscribe.
func NewDashboardMonitor(conn *websocket.Conn, wsMu *sync.Mutex, log logger.Logger, organizationID string, deployService DeployServiceProvider) (*DashboardMonitor, error) {
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	ctx := context.WithValue(context.Background(), shared_types.OrganizationIDKey, orgID.String())

	manager, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH manager: %w", err)
	}

	dockerService, err := getDockerService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker service: %w", err)
	}

	poller := getOrCreateOrgPoller(organizationID, manager, dockerService, deployService, log)

	monitor := &DashboardMonitor{
		conn:       conn,
		connMutex:  wsMu,
		log:        log,
		poller:     poller,
		Interval:   defaultPollerInterval,
		Operations: AllOperations,
	}

	return monitor, nil
}

// Start subscribes this monitor to the shared org poller. Safe to call
// multiple times; only the first call takes effect.
func (m *DashboardMonitor) Start() {
	m.subMu.Lock()
	defer m.subMu.Unlock()
	if m.subscribed {
		return
	}
	m.poller.subscribe(m)
	m.subscribed = true
}

// Stop unsubscribes from the poller and closes the WebSocket connection.
func (m *DashboardMonitor) Stop() {
	m.subMu.Lock()
	if m.subscribed {
		m.poller.unsubscribe(m)
		m.subscribed = false
	}
	m.subMu.Unlock()
	m.Close()
}

func (m *DashboardMonitor) Close() {
	m.connMutex.Lock()
	defer m.connMutex.Unlock()
	m.conn = nil
}

// UpdateConfig and SetOperations are kept for API compatibility.
// The shared poller always runs all operations at the default interval.
func (m *DashboardMonitor) UpdateConfig(config MonitoringConfig) {
	m.Interval = config.Interval
	if len(config.Operations) > 0 {
		m.Operations = config.Operations
	}
}

func (m *DashboardMonitor) SetOperations(operations []DashboardOperation) {
	m.Operations = operations
}
