package mover

import (
	"sync"
	"time"
)

// AppSessionInfo holds information about an app session for display
type AppSessionInfo struct {
	Name           string
	ApplicationID  string
	Status         ConnectionStatus
	FilesSynced    int
	ChangesDetected int
	URL            string
	Deployment     *DeploymentInfo
	Error          error
	Uptime         time.Duration
}

// MultiAppTracker manages status for multiple app sessions
type MultiAppTracker struct {
	mu        sync.RWMutex
	sessions  map[string]*AppSessionInfo
	startTime time.Time
}

// NewMultiAppTracker creates a new multi-app tracker
func NewMultiAppTracker() *MultiAppTracker {
	return &MultiAppTracker{
		sessions:  make(map[string]*AppSessionInfo),
		startTime: time.Now(),
	}
}

// UpdateSession updates the status of a specific app session
func (m *MultiAppTracker) UpdateSession(name string, info AppSessionInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info.Uptime = time.Since(m.startTime)
	m.sessions[name] = &info
}

// GetSessions returns a snapshot of all sessions
func (m *MultiAppTracker) GetSessions() []*AppSessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*AppSessionInfo, 0, len(m.sessions))
	for _, session := range m.sessions {
		info := *session
		info.Uptime = time.Since(m.startTime)
		result = append(result, &info)
	}
	return result
}

// GetUptime returns the uptime duration
func (m *MultiAppTracker) GetUptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Since(m.startTime)
}
