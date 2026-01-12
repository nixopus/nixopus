package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	healthcheck_service "github.com/raghavyuva/nixopus-api/internal/features/healthcheck/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/realtime"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type HealthCheckScheduler struct {
	service      *healthcheck_service.HealthCheckService
	logger       logger.Logger
	ctx          context.Context
	ticker       *time.Ticker
	stopChan     chan struct{}
	wg           sync.WaitGroup
	running      bool
	mu           sync.RWMutex
	socketServer *realtime.SocketServer
}

func NewHealthCheckScheduler(
	healthCheckService *healthcheck_service.HealthCheckService,
	logger logger.Logger,
	ctx context.Context,
) *HealthCheckScheduler {
	return &HealthCheckScheduler{
		service:      healthCheckService,
		logger:       logger,
		ctx:          ctx,
		stopChan:     make(chan struct{}),
		running:      false,
		socketServer: nil,
	}
}

// SetSocketServer sets the WebSocket server for broadcasting health check results
func (s *HealthCheckScheduler) SetSocketServer(socketServer *realtime.SocketServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.socketServer = socketServer
}

// Start begins the health check scheduler
func (s *HealthCheckScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.ticker = time.NewTicker(10 * time.Second)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		s.logger.Log(logger.Info, "Health check scheduler started", "")

		for {
			select {
			case <-s.ticker.C:
				s.runDueChecks()
			case <-s.stopChan:
				s.logger.Log(logger.Info, "Health check scheduler stopped", "")
				return
			}
		}
	}()
}

// Stop gracefully stops the health check scheduler
func (s *HealthCheckScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
	s.wg.Wait()
}

// runDueChecks executes health checks that are due
func (s *HealthCheckScheduler) runDueChecks() {
	checks, err := s.service.GetDueHealthChecks()
	if err != nil {
		s.logger.Log(logger.Error, "failed to get due health checks", err.Error())
		return
	}

	if len(checks) == 0 {
		return
	}

	s.logger.Log(logger.Info, "executing health checks", "count: "+fmt.Sprintf("%d", len(checks)))

	for _, check := range checks {
		go s.executeCheck(check)
	}
}

// executeCheck executes a single health check
func (s *HealthCheckScheduler) executeCheck(healthCheck *shared_types.HealthCheck) {
	result, err := s.service.ExecuteHealthCheck(healthCheck)
	if err != nil {
		s.logger.Log(logger.Error, "failed to execute health check", err.Error())
		return
	}

	if err := s.service.ProcessHealthCheckResult(healthCheck, result); err != nil {
		s.logger.Log(logger.Error, "failed to process health check result", err.Error())
		return
	}

	status := result.Status
	if status == string(shared_types.HealthCheckStatusHealthy) {
		s.logger.Log(logger.Info, "health check passed", "application_id: "+healthCheck.ApplicationID.String())
	} else {
		s.logger.Log(logger.Error, "health check failed", "application_id: "+healthCheck.ApplicationID.String()+", status: "+status)
	}

	// Broadcast result via WebSocket
	s.broadcastHealthCheckResult(healthCheck, result)
}

// broadcastHealthCheckResult broadcasts health check result to subscribed clients
func (s *HealthCheckScheduler) broadcastHealthCheckResult(healthCheck *shared_types.HealthCheck, result *shared_types.HealthCheckResult) {
	s.mu.RLock()
	socketServer := s.socketServer
	s.mu.RUnlock()

	if socketServer == nil {
		s.logger.Log(logger.Warning, "SocketServer not set, cannot broadcast health check result", "application_id: "+healthCheck.ApplicationID.String())
		return
	}

	if result == nil {
		s.logger.Log(logger.Error, "Health check result is nil, cannot broadcast", "application_id: "+healthCheck.ApplicationID.String())
		return
	}

	applicationID := healthCheck.ApplicationID.String()
	payload := map[string]interface{}{
		"application_id":    applicationID,
		"health_check_id":   healthCheck.ID.String(),
		"status":            result.Status,
		"response_time_ms":  result.ResponseTimeMs,
		"checked_at":        result.CheckedAt.Format(time.RFC3339),
		"consecutive_fails": healthCheck.ConsecutiveFails,
	}

	// Only include status_code if it's set (non-zero)
	if result.StatusCode != 0 {
		payload["status_code"] = result.StatusCode
	}

	// Only include error_message if it's not empty
	if result.ErrorMessage != "" {
		payload["error_message"] = result.ErrorMessage
	}

	// Validate payload is not empty
	if len(payload) == 0 {
		s.logger.Log(logger.Error, "Payload is empty, cannot broadcast", "application_id: "+applicationID)
		return
	}

	socketServer.BroadcastToTopic(realtime.MonitorHealthCheck, applicationID, payload)
	s.logger.Log(logger.Info, "Broadcasted health check result", "application_id: "+applicationID+", status: "+result.Status)
}
