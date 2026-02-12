package live

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/swarm"
	"github.com/google/uuid"
	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// ServiceManager handles dev service lifecycle management
type ServiceManager struct {
	stagingManager       *StagingManager
	taskService          *tasks.TaskService
	logger               logger.Logger
	startingMu           sync.Mutex
	startingApplications map[uuid.UUID]bool // Track applications that are starting
	// serviceReady tracks applications that already have a healthy service running,
	// preventing repeated Docker API calls on every file write.
	serviceReadyMu sync.RWMutex
	serviceReady   map[uuid.UUID]bool
	// ensureDebounceMu protects the debounce timers
	ensureDebounceMu sync.Mutex
	ensureTimers     map[uuid.UUID]*time.Timer
}

func NewServiceManager(stagingManager *StagingManager, taskService *tasks.TaskService, logger logger.Logger) *ServiceManager {
	return &ServiceManager{
		stagingManager:       stagingManager,
		taskService:          taskService,
		logger:               logger,
		startingApplications: make(map[uuid.UUID]bool),
		serviceReady:         make(map[uuid.UUID]bool),
		ensureTimers:         make(map[uuid.UUID]*time.Timer),
	}
}

// EnsureDevServiceStarted ensures the dev service is running for the application.
// This method is debounced: rapid calls (e.g. during initial file sync) are coalesced
// into a single check after a short delay, preventing a flood of SSH tunnel creations.
func (sm *ServiceManager) EnsureDevServiceStarted(ctx context.Context, appCtx *ApplicationContext) {
	// Fast path: if we already know the service is healthy, skip entirely.
	// The bind mount picks up file changes automatically once the service is running.
	sm.serviceReadyMu.RLock()
	ready := sm.serviceReady[appCtx.ApplicationID]
	sm.serviceReadyMu.RUnlock()
	if ready {
		return
	}

	// Debounce: coalesce rapid calls into a single check after 2 seconds.
	// During initial sync, dozens of files arrive per second — we only need one check.
	sm.ensureDebounceMu.Lock()
	if existing, ok := sm.ensureTimers[appCtx.ApplicationID]; ok {
		existing.Stop()
	}
	// Capture context values we need (context itself may be cancelled by the time timer fires)
	ctxCopy := context.WithoutCancel(ctx)
	sm.ensureTimers[appCtx.ApplicationID] = time.AfterFunc(2*time.Second, func() {
		sm.doEnsureDevServiceStarted(ctxCopy, appCtx)
	})
	sm.ensureDebounceMu.Unlock()
}

// doEnsureDevServiceStarted is the actual implementation that checks and starts the dev service.
func (sm *ServiceManager) doEnsureDevServiceStarted(ctx context.Context, appCtx *ApplicationContext) {
	// Clean up the timer reference
	sm.ensureDebounceMu.Lock()
	delete(sm.ensureTimers, appCtx.ApplicationID)
	sm.ensureDebounceMu.Unlock()

	// Find existing service by application ID
	existingService, err := tasks.FindServiceByLabel(ctx, "com.application.id", appCtx.ApplicationID.String())
	if err != nil {
		sm.logger.Log(logger.Error, "failed to check service existence by application ID", fmt.Sprintf("application_id=%s err=%v", appCtx.ApplicationID, err))
		return
	}

	// If service exists and is healthy, bind mount will automatically pick up changes
	if existingService != nil {
		if sm.isServiceHealthy(ctx, existingService) {
			// Mark as ready so future file writes skip the check entirely
			sm.serviceReadyMu.Lock()
			sm.serviceReady[appCtx.ApplicationID] = true
			sm.serviceReadyMu.Unlock()
			// Add domain with TLS if configured
			sm.addDomainWithTLSIfConfigured(ctx, existingService, appCtx)
			return
		}
		sm.logger.Log(logger.Info, "service exists but not running properly", appCtx.ApplicationID.String())
	}

	// Mark application as starting to prevent duplicate service creation attempts
	// (mutex-protected check - if another goroutine is already creating, return early)
	if !sm.markApplicationStarting(appCtx.ApplicationID) {
		return
	}

	defer sm.scheduleApplicationStartingClear(appCtx.ApplicationID)

	sm.logger.Log(logger.Info, "service not found, attempting service creation", appCtx.ApplicationID.String())

	cfg := sm.buildLiveDevConfig(appCtx)
	sm.logger.Log(logger.Info, "starting dev service", fmt.Sprintf("application_id=%s staging=%s domain=%s framework=%s", appCtx.ApplicationID, cfg.StagingPath, cfg.Domain, cfg.Framework))

	if err := sm.taskService.StartLiveDevTask(ctx, cfg); err != nil {
		sm.logger.Log(logger.Error, "failed to start dev service", fmt.Sprintf("application_id=%s err=%v", appCtx.ApplicationID, err))
		return
	}

	sm.logger.Log(logger.Info, "dev service task queued", appCtx.ApplicationID.String())
}

func (sm *ServiceManager) markApplicationStarting(applicationID uuid.UUID) bool {
	sm.startingMu.Lock()
	defer sm.startingMu.Unlock()
	if sm.startingApplications[applicationID] {
		return false
	}
	sm.startingApplications[applicationID] = true
	return true
}

func (sm *ServiceManager) scheduleApplicationStartingClear(applicationID uuid.UUID) {
	go func() {
		// Use a longer cooldown to prevent rapid retry loops when service creation
		// fails (e.g. framework detection failure). The debounced EnsureDevServiceStarted
		// will queue another attempt after this cooldown expires.
		time.Sleep(30 * time.Second)
		sm.startingMu.Lock()
		delete(sm.startingApplications, applicationID)
		sm.startingMu.Unlock()
	}()
}

// isServiceHealthy checks if the service is running with the desired number of tasks
func (sm *ServiceManager) isServiceHealthy(ctx context.Context, service *swarm.Service) bool {
	dockerService, err := docker.GetDockerServiceFromContext(ctx)
	if err != nil {
		return false
	}
	running, desired, err := dockerService.GetServiceHealth(*service)
	if err != nil || running < desired || desired == 0 {
		return false
	}
	return true
}

func (sm *ServiceManager) buildLiveDevConfig(appCtx *ApplicationContext) tasks.LiveDevConfig {
	cfg := tasks.LiveDevConfig{
		ApplicationID:  appCtx.ApplicationID,
		OrganizationID: appCtx.OrganizationID,
		StagingPath:    appCtx.StagingPath,
		Port:           0,
		EnvVars:        appCtx.EnvironmentVariables,
	}

	if appCtx.Domain != "" {
		cfg.Domain = appCtx.Domain
	} else if domain, ok := appCtx.Config["domain"].(string); ok && domain != "" {
		cfg.Domain = domain
	}

	if framework, ok := appCtx.Config["framework"].(string); ok && framework != "" {
		cfg.Framework = framework
	}

	return cfg
}

// addDomainWithTLSIfConfigured adds domain with TLS to Caddy if domain is configured
func (sm *ServiceManager) addDomainWithTLSIfConfigured(ctx context.Context, service *swarm.Service, appCtx *ApplicationContext) {
	// Get domain from appCtx
	domain := appCtx.Domain
	if domain == "" {
		if domainVal, ok := appCtx.Config["domain"].(string); ok && domainVal != "" {
			domain = domainVal
		}
	}

	if domain == "" {
		return
	}

	// Extract port from service endpoint spec
	port := sm.extractPortFromService(service)
	if port == 0 {
		sm.logger.Log(logger.Warning, "failed to extract port from service for domain", fmt.Sprintf("application_id=%s domain=%s", appCtx.ApplicationID, domain))
		return
	}

	// Add domain with TLS to Caddy
	client := caddygo.NewClient(config.AppConfig.Proxy.CaddyEndpoint)

	// Get SSH host from organization-specific SSH manager
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, appCtx.OrganizationID.String())
	manager, err := ssh.GetSSHManagerFromContext(orgCtx)
	if err != nil {
		sm.logger.Log(logger.Warning, "failed to get SSH manager", fmt.Sprintf("application_id=%s domain=%s err=%v", appCtx.ApplicationID, domain, err))
		return
	}
	upstreamHost, err := manager.GetSSHHost()
	if err != nil {
		sm.logger.Log(logger.Warning, "failed to get SSH host", fmt.Sprintf("application_id=%s domain=%s err=%v", appCtx.ApplicationID, domain, err))
		return
	}

	if err := client.AddDomainWithAutoTLS(domain, upstreamHost, port, caddygo.DomainOptions{}); err != nil {
		sm.logger.Log(logger.Warning, "failed to add domain with TLS to Caddy", fmt.Sprintf("application_id=%s domain=%s port=%d err=%v", appCtx.ApplicationID, domain, port, err))
		return
	}

	if err := client.Reload(); err != nil {
		sm.logger.Log(logger.Warning, "failed to reload Caddy after adding domain", fmt.Sprintf("application_id=%s domain=%s err=%v", appCtx.ApplicationID, domain, err))
		return
	}

	sm.logger.Log(logger.Info, "domain added with TLS successfully", fmt.Sprintf("application_id=%s domain=%s port=%d", appCtx.ApplicationID, domain, port))
}

// extractPortFromService extracts the published port from a swarm service
func (sm *ServiceManager) extractPortFromService(service *swarm.Service) int {
	if service == nil || len(service.Endpoint.Ports) == 0 {
		return 0
	}

	// Get the first published port
	for _, portConfig := range service.Endpoint.Ports {
		if portConfig.PublishedPort > 0 {
			return int(portConfig.PublishedPort)
		}
	}

	return 0
}
