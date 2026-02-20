package mover

import (
	"sync"
	"time"
)

// Tracker manages the status state for live development session
// Single Responsibility: State management only
type Tracker struct {
	mu sync.RWMutex

	connectionStatus ConnectionStatus
	serviceStatus    ServiceStatus
	filesSynced      int
	changesDetected  int
	url              string
	envPath          string
	deployment       *DeploymentInfo
	startTime        time.Time
}

// NewTracker creates a new status tracker
func NewTracker() *Tracker {
	return &Tracker{
		connectionStatus: ConnectionStatusDisconnected,
		serviceStatus:    ServiceStatusUnknown,
		startTime:        time.Now(),
	}
}

// SetConnectionStatus updates the connection status
func (t *Tracker) SetConnectionStatus(status ConnectionStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connectionStatus = status
}

// GetConnectionStatus returns the current connection status
func (t *Tracker) GetConnectionStatus() ConnectionStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connectionStatus
}

// SetServiceStatus updates the service status
func (t *Tracker) SetServiceStatus(status ServiceStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.serviceStatus = status
}

// GetServiceStatus returns the current service status
func (t *Tracker) GetServiceStatus() ServiceStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.serviceStatus
}

// IncrementFilesSynced increments the files synced counter
func (t *Tracker) IncrementFilesSynced() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.filesSynced++
}

// AddFilesSynced adds to the files synced counter
func (t *Tracker) AddFilesSynced(count int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.filesSynced += count
}

// GetFilesSynced returns the number of files synced
func (t *Tracker) GetFilesSynced() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.filesSynced
}

// IncrementChanges increments the changes detected counter
func (t *Tracker) IncrementChanges() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.changesDetected++
}

// GetChangesDetected returns the number of changes detected
func (t *Tracker) GetChangesDetected() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.changesDetected
}

// SetURL sets the application URL
func (t *Tracker) SetURL(url string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.url = url
}

// GetURL returns the application URL
func (t *Tracker) GetURL() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.url
}

// SetEnvPath sets the environment file path
func (t *Tracker) SetEnvPath(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.envPath = path
}

// GetEnvPath returns the environment file path
func (t *Tracker) GetEnvPath() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.envPath
}

// SetDeploymentInfo updates the deployment status information
func (t *Tracker) SetDeploymentInfo(info *DeploymentInfo) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.deployment = info
}

// GetDeploymentInfo returns the current deployment status
func (t *Tracker) GetDeploymentInfo() *DeploymentInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.deployment
}

// GetStatusInfo returns a snapshot of current status for rendering
func (t *Tracker) GetStatusInfo() StatusInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return StatusInfo{
		ConnectionStatus: t.connectionStatus,
		FilesSynced:      t.filesSynced,
		ChangesDetected:  t.changesDetected,
		ServiceStatus:    t.serviceStatus,
		URL:              t.url,
		EnvPath:          t.envPath,
		Deployment:       t.deployment,
	}
}

// GetUptime returns the uptime duration
func (t *Tracker) GetUptime() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return time.Since(t.startTime)
}
