package realtime

import (
	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/terminal"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// handleTerminal handles the terminal connection.
// It creates a new terminal if it doesn't exist, otherwise it writes the message to the existing terminal.
// The terminal connects to the specified server or the active server for the user's organization.
// Parameters:
//
//	conn - the *websocket.Conn representing the client connection.
//	msg - the types.Payload representing the message from the client.
//
// Expected payload data:
//
//	terminalId - unique identifier for the terminal session
//	value - input to write to the terminal
//	serverId - (optional) specific server ID to connect to
func (s *SocketServer) handleTerminal(conn *websocket.Conn, msg types.Payload) {
	s.terminalMutex.Lock()
	defer s.terminalMutex.Unlock()

	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, "Invalid terminal data")
		return
	}
	terminalId, ok := dataMap["terminalId"].(string)
	if !ok {
		s.sendError(conn, "Missing terminalId")
		return
	}
	input, ok := dataMap["value"].(string)
	if !ok {
		s.sendError(conn, "Invalid terminal input")
		return
	}

	// Get optional serverId - if provided, use this specific server
	var serverID string
	if sid, ok := dataMap["serverId"].(string); ok {
		serverID = sid
	}

	// Get user from connection to retrieve organization ID
	userAny, ok := s.conns.Load(conn)
	if !ok {
		s.sendError(conn, "User not found for connection")
		return
	}
	user, ok := userAny.(*types.User)
	if !ok {
		s.sendError(conn, "Invalid user data")
		return
	}

	// Get the user's first organization ID
	if len(user.Organizations) == 0 {
		s.sendError(conn, "User has no organizations")
		return
	}
	organizationID := user.Organizations[0].ID

	// Ensure map exists for this connection
	if s.terminals[conn] == nil {
		s.terminals[conn] = make(map[string]*terminal.Terminal)
	}

	term, exists := s.terminals[conn][terminalId]
	if !exists {
		log := logger.NewLogger()
		config := terminal.TerminalConfig{
			Conn:           conn,
			Log:            &log,
			TerminalId:     terminalId,
			DB:             s.db,
			Ctx:            s.ctx,
			OrganizationID: organizationID,
			ServerID:       serverID,
		}

		newTerminal, err := terminal.NewTerminalWithConfig(config)
		if err != nil {
			s.sendError(conn, "Failed to start terminal: "+err.Error())
			return
		}
		s.terminals[conn][terminalId] = newTerminal
		go newTerminal.Start()
		term = newTerminal
	}

	term.WriteMessage(input)
}

// handleTerminalResize handles the terminal resize.
// It resizes the terminal if it exists, otherwise it sends an error to the client.
// Parameters:
//
//	conn - the *websocket.Conn representing the client connection.
//	msg - the types.Payload representing the message from the client.
func (s *SocketServer) handleTerminalResize(conn *websocket.Conn, msg types.Payload) {
	s.terminalMutex.Lock()
	defer s.terminalMutex.Unlock()

	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, "Invalid resize data")
		return
	}

	terminalId, ok := data["terminalId"].(string)
	if !ok {
		s.sendError(conn, "Missing terminalId")
		return
	}

	term, exists := s.terminals[conn][terminalId]
	if !exists {
		s.sendError(conn, "Terminal not started")
		return
	}

	rows, ok := data["rows"].(float64)
	if !ok {
		s.sendError(conn, "Invalid rows value")
		return
	}

	cols, ok := data["cols"].(float64)
	if !ok {
		s.sendError(conn, "Invalid cols value")
		return
	}

	term.ResizeTerminal(uint16(rows), uint16(cols))
}
