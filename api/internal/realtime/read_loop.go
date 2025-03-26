package realtime

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/dashboard"
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
			if user == nil {
				s.sendError(conn, "Authentication required")
				continue
			}

			if msg.Topic != "" && msg.Data != nil {
				var resourceID string
				if dataMap, ok := msg.Data.(map[string]interface{}); ok {
					resourceID, ok = dataMap["resource_id"].(string)
					if !ok {
						s.sendError(conn, "Invalid topic subscription format. Requires resourceId")
						continue
					}
				}

				s.SubscribeToTopic(topics(msg.Topic), resourceID, conn)
				continue
			}
			s.sendError(conn, "Invalid topic subscription format")

		case "unsubscribe":
			if user == nil {
				s.sendError(conn, "Authentication required")
				continue
			}

			if msg.Topic != "" && msg.Data != nil {
				var resourceID string
				if dataMap, ok := msg.Data.(map[string]interface{}); ok {
					resourceID, ok = dataMap["resource_id"].(string)
					if !ok {
						s.sendError(conn, "Invalid topic unsubscription format. Requires resourceId")
						continue
					}
				}

				s.UnsubscribeFromTopic(topics(msg.Topic), resourceID, conn)
				continue
			}

			s.sendError(conn, "Invalid topic unsubscription format")

		case "authenticate":
			token, ok := msg.Data.(string)
			if !ok {
				s.sendError(conn, "Invalid authentication token format")
				continue
			}

			newUser, err := s.verifyToken(token)
			if err != nil {
				s.sendError(conn, "Invalid authorization token")
				continue
			}

			user = newUser
			s.conns.Store(conn, user.ID)

			conn.WriteJSON(types.Payload{
				Action: "authenticated",
				Data:   user.ID,
			})

			log.Printf("User re-authenticated. ID: %s, Email: %s", user.ID, user.Email)

		case "terminal":
			if user == nil {
				s.sendError(conn, "Authentication required")
				continue
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
					continue
				}

				s.terminalMutex.Lock()
				s.terminals[conn] = newTerminal
				s.terminalMutex.Unlock()

				newTerminal.WriteMessage(msg.Data.(string))

				go newTerminal.Start()
			}

		case "dashboard_monitor":
			if user == nil {
				s.sendError(conn, "Authentication required")
				continue
			}

			s.dashboardMutex.Lock()
			monitor, exists := s.dashboardMonitors[conn]
			if !exists {
				newMonitor, err := dashboard.NewDashboardMonitor(conn, logger.NewLogger())
				if err != nil {
					s.dashboardMutex.Unlock()
					s.sendError(conn, "Failed to create dashboard monitor")
					continue
				}

				s.dashboardMonitors[conn] = newMonitor
				monitor = newMonitor
			}
			s.dashboardMutex.Unlock()

			if msg.Data != nil {
				dataMap, ok := msg.Data.(map[string]interface{})
				if !ok {
					s.sendError(conn, "Invalid dashboard monitor configuration")
					continue
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
					continue
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
					continue
				}

				conn.WriteMessage(websocket.TextMessage, jsonData)
			}
		case "stop_dashboard_monitor":
			if user == nil {
				s.sendError(conn, "Authentication required")
				continue
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

		default:
			s.sendError(conn, "Unknown message action")
		}
	}
}
