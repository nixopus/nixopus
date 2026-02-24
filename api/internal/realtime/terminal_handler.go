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

	term := s.getOrCreateTerminal(conn, terminalId)
	if term == nil {
		return
	}

	if err := term.WriteMessage(input); err != nil {
		fmt.Printf("[ws] handleTerminal: WriteMessage failed for terminal %s: %v\n", terminalId, err)
	}
}

// getOrCreateTerminal returns the existing terminal for the given ID or creates
// a new one. The mutex is only held during map lookup/creation, never during
// blocking I/O. Returns nil if terminal creation failed (error sent to client).
func (s *SocketServer) getOrCreateTerminal(conn *websocket.Conn, terminalId string) *terminal.Terminal {
	s.terminalMutex.Lock()
	defer s.terminalMutex.Unlock()

	if s.terminals[conn] == nil {
		s.terminals[conn] = make(map[string]*terminal.Terminal)
	}

	if term, exists := s.terminals[conn][terminalId]; exists {
		if !term.IsDone() {
			return term
		}
		fmt.Printf("[ws] getOrCreateTerminal: terminal %s is dead, cleaning up and recreating\n", terminalId)
		term.Close()
		delete(s.terminals[conn], terminalId)
	}

	return s.createTerminal(conn, terminalId)
}

func (s *SocketServer) createTerminal(conn *websocket.Conn, terminalId string) *terminal.Terminal {
	fmt.Printf("[ws] createTerminal: creating terminal %s\n", terminalId)

	orgIDVal, ok := s.orgIDs.Load(conn)
	if !ok || orgIDVal == nil {
		s.sendError(conn, "Organization ID not found for this connection")
		return nil
	}

	orgIDStr, ok := orgIDVal.(string)
	if !ok || orgIDStr == "" {
		s.sendError(conn, "Invalid organization ID for this connection")
		return nil
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		s.sendError(conn, fmt.Sprintf("Invalid organization ID format: %v", err))
		return nil
	}

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	log := logger.NewLogger()
	newTerminal, err := terminal.NewTerminal(ctx, conn, s.getConnWriteMu(conn), &log, terminalId)
	if err != nil {
		fmt.Printf("[ws] createTerminal: failed to create terminal %s: %v\n", terminalId, err)
		s.sendError(conn, fmt.Sprintf("Failed to start terminal: %v", err))
		return nil
	}
	s.terminals[conn][terminalId] = newTerminal
	go newTerminal.Start()
	fmt.Printf("[ws] createTerminal: terminal %s started\n", terminalId)
	return newTerminal
}

// handleTerminalResize handles the terminal resize.
// It resizes the terminal if it exists, otherwise it sends an error to the client.
// Parameters:
//
//	conn - the *websocket.Conn representing the client connection.
//	msg - the types.Payload representing the message from the client.
func (s *SocketServer) handleTerminalResize(conn *websocket.Conn, msg types.Payload) {
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

	s.terminalMutex.RLock()
	term, exists := s.terminals[conn][terminalId]
	s.terminalMutex.RUnlock()

	if !exists {
		s.sendError(conn, "Terminal not started")
		return
	}

	term.ResizeTerminal(uint16(rows), uint16(cols))
}
