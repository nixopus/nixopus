package realtime

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/dashboard"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/realtime"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/terminal"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *SocketServer) readLoop(conn *websocket.Conn, user *types.User) {
	defer func() {
		s.terminalMutex.Lock()
		if term, exists := s.terminals[conn]; exists {
			term.Close()
			delete(s.terminals, conn)
		}
		s.terminalMutex.Unlock()

		s.dashboardMutex.Lock()
		if monitor, exists := s.dashboardMonitors[conn]; exists {
			monitor.Stop()
			delete(s.dashboardMonitors, conn)
		}
		s.dashboardMutex.Unlock()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading message: %v\n", err)
			s.sendError(conn, "Failed to read message")
			return
		}

		var msg types.Payload
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			s.sendError(conn, "Invalid message format")
			continue
		}

		switch msg.Action {
		case "ping":
			s.handlePing(conn)

		case "subscribe":
			s.handleSubscribe(conn, msg, user)

		case "unsubscribe":
			s.handleUnsubscribe(conn, msg, user)

		case "authenticate":
			user = s.handleAuthenticate(conn, msg)

		case "terminal":
			s.handleTerminal(conn, msg, user)

		case "terminal_resize":
			s.handleTerminalResize(conn, msg, user)

		case "dashboard_monitor":
			s.handleDashboardMonitor(conn, msg, user)

		case "stop_dashboard_monitor":
			s.handleStopDashboardMonitor(conn, user)
		case "monitor_application":
			s.handleMonitorApplication(conn, msg, user)

		default:
			s.sendError(conn, "Unknown message action")
		}
	}
}

func (s *SocketServer) handleSubscribe(conn *websocket.Conn, msg types.Payload, user *types.User) {
	if user == nil {
		s.sendError(conn, "Authentication required")
		return
	}

	if msg.Topic != "" && msg.Data != nil {
		var resourceID string
		if dataMap, ok := msg.Data.(map[string]interface{}); ok {
			resourceID, ok = dataMap["resource_id"].(string)
			if !ok {
				s.sendError(conn, "Invalid topic subscription format. Requires resourceId")
				return
			}
		}

		s.SubscribeToTopic(topics(msg.Topic), resourceID, conn)
		return
	}
	s.sendError(conn, "Invalid topic subscription format")
}

func (s *SocketServer) handleUnsubscribe(conn *websocket.Conn, msg types.Payload, user *types.User) {
	if user == nil {
		s.sendError(conn, "Authentication required")
		return
	}

	if msg.Topic != "" && msg.Data != nil {
		var resourceID string
		if dataMap, ok := msg.Data.(map[string]interface{}); ok {
			resourceID, ok = dataMap["resource_id"].(string)
			if !ok {
				s.sendError(conn, "Invalid topic unsubscription format. Requires resourceId")
				return
			}
		}

		s.UnsubscribeFromTopic(topics(msg.Topic), resourceID, conn)
		return
	}

	s.sendError(conn, "Invalid topic unsubscription format")
}

func (s *SocketServer) handleAuthenticate(conn *websocket.Conn, msg types.Payload) *types.User {
	token, ok := msg.Data.(string)
	if !ok {
		s.sendError(conn, "Invalid authentication token format")
		return nil
	}

	newUser, err := s.verifyToken(token)
	if err != nil {
		s.sendError(conn, "Invalid authorization token")
		return nil
	}

	s.conns.Store(conn, newUser.ID)

	conn.WriteJSON(types.Payload{
		Action: "authenticated",
		Data:   newUser.ID,
	})

	log.Printf("User re-authenticated. ID: %s, Email: %s", newUser.ID, newUser.Email)

	return newUser
}

func (s *SocketServer) handleTerminal(conn *websocket.Conn, msg types.Payload, user *types.User) {
	if user == nil {
		s.sendError(conn, "Authentication required")
		return
	}
	s.terminalMutex.Lock()
	term, exists := s.terminals[conn]
	s.terminalMutex.Unlock()

	if exists {
		s.terminalMutex.Lock()
		term.WriteMessage(msg.Data.(string))
		s.terminalMutex.Unlock()
	} else {
		newTerminal, err := terminal.NewTerminal(conn, &logger.Logger{})
		if err != nil {
			s.sendError(conn, "Failed to start terminal")
			return
		}

		s.terminalMutex.Lock()
		s.terminals[conn] = newTerminal
		s.terminalMutex.Unlock()

		newTerminal.WriteMessage(msg.Data.(string))

		go newTerminal.Start()
	}
}

func (s *SocketServer) handleTerminalResize(conn *websocket.Conn, msg types.Payload, user *types.User) {
	if user == nil {
		s.sendError(conn, "Authentication required")
		return
	}
	s.terminalMutex.Lock()
	term, exists := s.terminals[conn]
	s.terminalMutex.Unlock()

	if exists {
		s.terminalMutex.Lock()
		rows := uint16(msg.Data.(map[string]interface{})["rows"].(float64))
		cols := uint16(msg.Data.(map[string]interface{})["cols"].(float64))
		term.ResizeTerminal(rows, cols)
		s.terminalMutex.Unlock()
	} else {
		s.sendError(conn, "Terminal not started")
	}
}

func (s *SocketServer) handleMonitorApplication(conn *websocket.Conn, msg types.Payload, user *types.User) {
	if user == nil {
		s.sendError(conn, "Authentication required")
		return
	}

	s.applicationMutex.Lock()
	monitor, exists := s.applicationMonitors[conn]
	if !exists {
		newMonitor, err := realtime.NewApplicationMonitor(conn, logger.NewLogger())
		if err != nil {
			s.applicationMutex.Unlock()
			s.sendError(conn, "Failed to create application monitor")
			return
		}

		s.applicationMonitors[conn] = newMonitor
		monitor = newMonitor
	}
	s.applicationMutex.Unlock()

	if msg.Data != nil {
		dataMap, ok := msg.Data.(map[string]interface{})
		if !ok {
			s.sendError(conn, "Invalid application monitor configuration")
			return
		}

		var interval time.Duration
		if intervalSec, ok := dataMap["interval"].(float64); ok {
			interval = time.Duration(intervalSec) * time.Second
		} else {
			interval = 10 * time.Second
		}

		var operations []realtime.ApplicationMonitorOperation
		if ops, ok := dataMap["operations"].([]interface{}); ok {
			for _, op := range ops {
				if opStr, ok := op.(string); ok {
					operations = append(operations, realtime.ApplicationMonitorOperation(opStr))
				}
			}
		}

		if len(operations) == 0 {
			operations = []realtime.ApplicationMonitorOperation{
				realtime.ContainerStatistics,
			}
		}

		config := realtime.MonitoringConfig{
			Interval:   interval,
			Operations: operations,
		}

		monitor.Interval = config.Interval
		monitor.Operations = config.Operations

		monitor.Start()

		response := types.Payload{
			Action: "application_monitor_started",
			Data: map[string]interface{}{
				"interval":   interval.Seconds(),
				"operations": operations,
			},
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			s.sendError(conn, "Failed to marshal response")
			return
		}

		conn.WriteMessage(websocket.TextMessage, jsonData)
	} else {
		monitor.Stop()
		response := types.Payload{
			Action: "application_monitor_stopped",
			Data:   nil,
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			s.sendError(conn, "Failed to marshal response")
			return
		}

		conn.WriteMessage(websocket.TextMessage, jsonData)
	}
}

func (s *SocketServer) handleStopDashboardMonitor(conn *websocket.Conn, user *types.User) {
	if user == nil {
		s.sendError(conn, "Authentication required")
		return
	}

	s.dashboardMutex.Lock()
	if monitor, exists := s.dashboardMonitors[conn]; exists {
		monitor.Stop()
		delete(s.dashboardMonitors, conn)

		conn.WriteJSON(types.Payload{
			Action: "dashboard_monitor_stopped",
			Data:   nil,
		})
	}
	s.dashboardMutex.Unlock()
}

func (s *SocketServer) handleDashboardMonitor(conn *websocket.Conn, msg types.Payload, user *types.User) {
	if user == nil {
		s.sendError(conn, "Authentication required")
		return
	}
	s.dashboardMutex.Lock()
	monitor, exists := s.dashboardMonitors[conn]
	if !exists {
		newMonitor, err := dashboard.NewDashboardMonitor(conn, logger.NewLogger())
		if err != nil {
			s.dashboardMutex.Unlock()
			s.sendError(conn, "Failed to create dashboard monitor")
			return
		}

		s.dashboardMonitors[conn] = newMonitor
		monitor = newMonitor
	}
	s.dashboardMutex.Unlock()

	if msg.Data != nil {
		dataMap, ok := msg.Data.(map[string]interface{})
		if !ok {
			s.sendError(conn, "Invalid dashboard monitor configuration")
			return
		}

		var interval time.Duration
		if intervalSec, ok := dataMap["interval"].(float64); ok {
			interval = time.Duration(intervalSec) * time.Second
		} else {
			interval = 10 * time.Second
		}

		var operations []dashboard.DashboardOperation
		if ops, ok := dataMap["operations"].([]interface{}); ok {
			for _, op := range ops {
				if opStr, ok := op.(string); ok {
					operations = append(operations, dashboard.DashboardOperation(opStr))
				}
			}
		}

		if len(operations) == 0 {
			operations = dashboard.AllOperations
		}

		config := dashboard.MonitoringConfig{
			Interval:   interval,
			Operations: operations,
		}

		monitor.Interval = config.Interval
		monitor.Operations = config.Operations

		monitor.Start()

		response := types.Payload{
			Action: "dashboard_monitor_started",
			Data: map[string]interface{}{
				"interval":   interval.Seconds(),
				"operations": operations,
			},
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			s.sendError(conn, "Failed to marshal response")
			return
		}

		conn.WriteMessage(websocket.TextMessage, jsonData)
	} else {
		monitor.Stop()
		response := types.Payload{
			Action: "dashboard_monitor_stopped",
			Data:   nil,
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			s.sendError(conn, "Failed to marshal response")
			return
		}

		conn.WriteMessage(websocket.TextMessage, jsonData)
	}
}
