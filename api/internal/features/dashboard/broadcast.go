package dashboard

import (
	"encoding/json"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (m *DashboardMonitor) Broadcast(action string, message interface{}) {
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

		if err := m.conn.WriteJSON(map[string]interface{}{"action": action, "data": message, "timestamp": time.Now().Unix(), "topic": "dashboard_monitor"}); err != nil {
			m.log.Log(logger.Error, "Failed to broadcast message", err.Error())
		}

		_ = m.conn.SetWriteDeadline(time.Time{})

	case <-time.After(3 * time.Second):
		m.log.Log(logger.Error, "Timeout waiting for broadcast lock", "")
	}
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
