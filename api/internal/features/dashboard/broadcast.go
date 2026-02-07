package dashboard

import (
	"encoding/json"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (m *DashboardMonitor) Broadcast(action string, message interface{}) {
	m.connMutex.Lock()
	defer m.connMutex.Unlock()

	// Check connection before any operations
	if m.conn == nil {
		return
	}

	// Prepare message data
	msg := map[string]interface{}{
		"action":    action,
		"data":      message,
		"timestamp": time.Now().Unix(),
		"topic":     "dashboard_monitor",
	}

	// Set write deadline
	deadline := time.Now().Add(5 * time.Second)
	if err := m.conn.SetWriteDeadline(deadline); err != nil {
		m.log.Log(logger.Error, "Failed to set write deadline", err.Error())
		return
	}

	// Write JSON message - this is the only place that writes to the connection
	// The mutex ensures only one goroutine can execute this at a time
	if err := m.conn.WriteJSON(msg); err != nil {
		m.log.Log(logger.Error, "Failed to broadcast message", err.Error())
		return
	}

	// Reset deadline after successful write
	_ = m.conn.SetWriteDeadline(time.Time{})
}

func (m *DashboardMonitor) BroadcastDebug(message string) {
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

func (m *DashboardMonitor) BroadcastError(errMsg string, operation DashboardOperation) {
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
