package realtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/terminal"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// handleTerminal handles the terminal connection.
// It creates a new terminal if it doesn't exist, otherwise it writes the message to the existing terminal.
// Parameters:
//
//	conn - the *websocket.Conn representing the client connection.
//	msg - the types.Payload representing the message from the client.
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

	// Ensure map exists for this connection
	if s.terminals[conn] == nil {
		s.terminals[conn] = make(map[string]*terminal.Terminal)
	}

	term, exists := s.terminals[conn][terminalId]
	if !exists {
		// Get organization ID from connection
		orgIDVal, ok := s.orgIDs.Load(conn)
		if !ok || orgIDVal == nil {
			s.sendError(conn, "Organization ID not found for this connection")
			return
		}

		orgIDStr, ok := orgIDVal.(string)
		if !ok || orgIDStr == "" {
			s.sendError(conn, "Invalid organization ID for this connection")
			return
		}

		// Parse and set organization ID in context
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			s.sendError(conn, fmt.Sprintf("Invalid organization ID format: %v", err))
			return
		}

		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
		newTerminal, err := terminal.NewTerminal(ctx, conn, &logger.Logger{}, terminalId)
		if err != nil {
			s.sendError(conn, fmt.Sprintf("Failed to start terminal: %v", err))
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
