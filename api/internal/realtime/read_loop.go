package realtime

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// readLoop is the main read loop for the socket server.
// It reads messages from the client and handles them. based on the action, it will call the appropriate handler.
// Parameters
//
//	conn - the *websocket.Conn representing the client connection.
//	user - the *types.User representing the authenticated user.
func (s *SocketServer) readLoop(conn *websocket.Conn) {
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
		case types.PING:
			s.handlePing(conn)

		case types.SUBSCRIBE:
			s.handleSubscribe(conn, msg)

		case types.UNSUBSCRIBE:
			s.handleUnsubscribe(conn, msg)

		case types.TERMINAL:
			s.handleTerminal(conn, msg)

		case types.TERMINAL_RESIZE:
			s.handleTerminalResize(conn, msg)

		case types.DASHBOARD_MONITOR:
			s.handleDashboardMonitor(conn, msg)

		case types.STOP_DASHBOARD_MONITOR:
			s.handleStopDashboardMonitor(conn)

		case types.MONITOR_APPLICATION:
			// s.handleMonitorApplication(conn, msg, user)

		default:
			s.sendError(conn, "Unknown message action")
		}
	}
}
