package caddy

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/vmihailenco/taskq/v3"
)

// ServerHealth tracks the health state of a remote server's Caddy instance.
type ServerHealth struct {
	OrganizationID     uuid.UUID
	Host               string
	Healthy            bool
	LastCheck          time.Time
	LastHealthy        time.Time
	FailCount          int
	LastError          string
	RestartAttempts    int
	LastRestartAttempt time.Time
	ContainerMissing   bool
}

const (
	restartBaseDelay = 1 * time.Minute
	restartMaxDelay  = 10 * time.Minute
)

// HealthMonitor distributes health checks across Redis-backed workers and
// triggers reconciliation when a Caddy instance recovers from a failure.
// The scheduler loop runs on one instance; the actual checks are processed
// by any consumer via the shared Redis queue.
type HealthMonitor struct {
	logger     logger.Logger
	reconciler *Reconciler
	interval   time.Duration
	orgFetcher func(ctx context.Context) ([]uuid.UUID, error)
	stopCh     chan struct{}

	mu      sync.RWMutex
	servers map[uuid.UUID]*ServerHealth
}

func NewHealthMonitor(
	lgr logger.Logger,
	reconciler *Reconciler,
	interval time.Duration,
	orgFetcher func(ctx context.Context) ([]uuid.UUID, error),
) *HealthMonitor {
	return &HealthMonitor{
		logger:     lgr,
		reconciler: reconciler,
		interval:   interval,
		orgFetcher: orgFetcher,
		stopCh:     make(chan struct{}),
		servers:    make(map[uuid.UUID]*ServerHealth),
	}
}

// SetupQueue registers the health-check task handler on the shared Redis queue.
// Must be called after ReconcilerDaemon.SetupQueues().
func (h *HealthMonitor) SetupQueue() {
	if TaskCaddyHealthCheck != nil {
		return
	}

	TaskCaddyHealthCheck = taskq.RegisterTask(&taskq.TaskOptions{
		Name:       TASK_CADDY_HEALTH,
		RetryLimit: 0,
		Handler: func(ctx context.Context, payload CaddyTaskPayload) error {
			orgID, err := uuid.Parse(payload.OrganizationID)
			if err != nil {
				return fmt.Errorf("invalid org ID in health check task: %w", err)
			}
			h.checkServer(ctx, orgID)
			return nil
		},
	})
}

func (h *HealthMonitor) Start(ctx context.Context) {
	go h.run(ctx)
}

func (h *HealthMonitor) Stop() {
	close(h.stopCh)
}

// GetServerHealth returns the current health state for an organization's server.
func (h *HealthMonitor) GetServerHealth(orgID uuid.UUID) *ServerHealth {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if s, ok := h.servers[orgID]; ok {
		cp := *s
		return &cp
	}
	return nil
}

// GetAllHealth returns a snapshot of all monitored servers' health.
func (h *HealthMonitor) GetAllHealth() []ServerHealth {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]ServerHealth, 0, len(h.servers))
	for _, s := range h.servers {
		result = append(result, *s)
	}
	return result
}

func (h *HealthMonitor) run(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			h.logger.Log(logger.Info, "health monitor stopped", "")
			return
		case <-ctx.Done():
			h.logger.Log(logger.Info, "health monitor context cancelled", "")
			return
		case <-ticker.C:
			h.enqueueAll(ctx)
		}
	}
}

// enqueueAll pushes a health-check job for each org into the Redis queue.
// The actual checks run on whichever consumer picks them up.
func (h *HealthMonitor) enqueueAll(ctx context.Context) {
	orgIDs, err := h.orgFetcher(ctx)
	if err != nil {
		h.logger.Log(logger.Error, "health monitor: failed to fetch organizations", err.Error())
		return
	}

	if HealthCheckQueue == nil || TaskCaddyHealthCheck == nil {
		h.logger.Log(logger.Warning, "health check queue not initialized, falling back to direct checks", "")
		for _, orgID := range orgIDs {
			h.checkServer(ctx, orgID)
		}
		return
	}

	for _, orgID := range orgIDs {
		msg := TaskCaddyHealthCheck.WithArgs(ctx, CaddyTaskPayload{
			OrganizationID: orgID.String(),
		})
		msg.OnceInPeriod(15 * time.Second)

		if err := HealthCheckQueue.Add(msg); err != nil {
			h.logger.Log(logger.Warning, "failed to enqueue health check for org "+orgID.String(), err.Error())
		}
	}
}

func (h *HealthMonitor) checkServer(ctx context.Context, orgID uuid.UUID) {
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, orgID.String())

	host, err := getSSHHostForOrg(orgCtx)
	if err != nil {
		h.updateHealth(orgID, "", false, fmt.Sprintf("ssh host lookup failed: %v", err))
		return
	}

	err = PingCaddy(orgCtx, nil, &h.logger)

	h.mu.RLock()
	prev, existed := h.servers[orgID]
	wasHealthy := existed && prev.Healthy
	h.mu.RUnlock()

	if err != nil {
		h.updateHealth(orgID, host, false, err.Error())

		health := h.GetServerHealth(orgID)
		if health != nil && health.FailCount >= 3 && h.shouldAttemptRestart(orgID) {
			h.logger.Log(logger.Warning,
				fmt.Sprintf("caddy on %s unreachable for %d consecutive checks, attempting container restart", host, health.FailCount),
				orgID.String())
			h.attemptCaddyRestart(orgCtx, orgID, host)
		}
		return
	}

	h.updateHealth(orgID, host, true, "")

	if !wasHealthy && existed {
		h.logger.Log(logger.Info,
			fmt.Sprintf("caddy on %s recovered, triggering reconciliation", host),
			orgID.String())
		h.triggerReconciliation(orgID)
	}
}

func (h *HealthMonitor) updateHealth(orgID uuid.UUID, host string, healthy bool, errMsg string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	s, exists := h.servers[orgID]
	if !exists {
		s = &ServerHealth{
			OrganizationID: orgID,
			Host:           host,
		}
		h.servers[orgID] = s
	}

	s.Host = host
	s.LastCheck = time.Now()
	s.LastError = errMsg

	if healthy {
		s.Healthy = true
		s.FailCount = 0
		s.LastHealthy = time.Now()
		s.RestartAttempts = 0
		s.ContainerMissing = false
		s.LastRestartAttempt = time.Time{}
	} else {
		s.Healthy = false
		s.FailCount++
	}
}

func (h *HealthMonitor) shouldAttemptRestart(orgID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	s, exists := h.servers[orgID]
	if !exists || s.RestartAttempts == 0 {
		return true
	}
	backoff := restartBaseDelay * time.Duration(1<<uint(min(s.RestartAttempts, 6)))
	if backoff > restartMaxDelay {
		backoff = restartMaxDelay
	}
	return time.Since(s.LastRestartAttempt) >= backoff
}

func (h *HealthMonitor) recordRestartAttempt(orgID uuid.UUID, containerMissing bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s, exists := h.servers[orgID]; exists {
		s.RestartAttempts++
		s.LastRestartAttempt = time.Now()
		s.ContainerMissing = containerMissing
	}
}

func (h *HealthMonitor) attemptCaddyRestart(ctx context.Context, orgID uuid.UUID, host string) {
	manager, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		h.logger.Log(logger.Error, "failed to get SSH manager for caddy restart", err.Error())
		return
	}

	sshClient, err := manager.GetDefaultSSH()
	if err != nil {
		h.logger.Log(logger.Error, "failed to get SSH config for caddy restart", err.Error())
		return
	}

	// Create a dedicated, non-pooled connection so that closing it after the
	// restart command doesn't kill other sessions (e.g. terminals) that share
	// the pooled connection.
	conn, err := sshClient.Connect()
	if err != nil {
		h.logger.Log(logger.Error, "failed to SSH connect for caddy restart", err.Error())
		return
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		h.logger.Log(logger.Error, "failed to create SSH session for caddy restart", err.Error())
		return
	}
	defer session.Close()

	output, err := session.CombinedOutput("docker restart nixopus-caddy")
	if err != nil {
		missing := strings.Contains(string(output), "No such container")
		h.recordRestartAttempt(orgID, missing)

		if missing {
			h.logger.Log(logger.Warning,
				fmt.Sprintf("caddy container does not exist on %s, will retry with backoff", host),
				orgID.String())
		} else {
			h.logger.Log(logger.Error,
				fmt.Sprintf("caddy container restart failed on %s", host),
				fmt.Sprintf("error: %v, output: %s", err, string(output)))
		}
		return
	}

	h.recordRestartAttempt(orgID, false)
	h.logger.Log(logger.Info, fmt.Sprintf("caddy container restarted on %s", host), orgID.String())

	InvalidateTunnel(host)
}

// triggerReconciliation enqueues a reconcile job via Redis instead of running
// it inline, so any available worker picks it up.
func (h *HealthMonitor) triggerReconciliation(orgID uuid.UUID) {
	if err := EnqueueReconcile(orgID); err != nil {
		h.logger.Log(logger.Error, "failed to enqueue post-recovery reconciliation", err.Error())
	}
}
