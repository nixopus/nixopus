package dashboard

import (
	"encoding/json"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// send writes a single message to this monitor's WebSocket connection.
func (m *DashboardMonitor) send(action string, message interface{}) {
	m.connMutex.Lock()
	defer m.connMutex.Unlock()

	if m.conn == nil {
		return
	}

	msg := map[string]interface{}{
		"action":    action,
		"data":      message,
		"timestamp": time.Now().Unix(),
		"topic":     "dashboard_monitor",
	}

	deadline := time.Now().Add(5 * time.Second)
	if err := m.conn.SetWriteDeadline(deadline); err != nil {
		m.log.Log(logger.Error, "Failed to set write deadline", err.Error())
		return
	}

	if err := m.conn.WriteJSON(msg); err != nil {
		m.log.Log(logger.Error, "Failed to send message", err.Error())
		return
	}

	_ = m.conn.SetWriteDeadline(time.Time{})
}

// sendError writes an error message to this monitor's WebSocket connection.
func (m *DashboardMonitor) sendError(errMsg string, operation DashboardOperation) {
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

	m.send("error", string(jsonData))
}

// broadcast fans out a message to every subscribed monitor.
func (p *OrgPoller) broadcast(action string, data interface{}) {
	p.subMu.RLock()
	defer p.subMu.RUnlock()
	for m := range p.subscribers {
		m.send(action, data)
	}
}

// broadcastError fans out an error to every subscribed monitor.
func (p *OrgPoller) broadcastError(errMsg string, op DashboardOperation) {
	p.subMu.RLock()
	defer p.subMu.RUnlock()
	for m := range p.subscribers {
		m.sendError(errMsg, op)
	}
}
