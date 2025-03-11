package realtime

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *SocketServer) waitForAuth(conn *websocket.Conn, authChan chan<- *types.User) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		fmt.Printf("Error reading auth message: %v\n", err)
		s.sendError(conn, "Failed to read authentication message")
		authChan <- nil
		return
	}

	var msg types.Payload
	if err := json.Unmarshal(message, &msg); err != nil {
		fmt.Printf("Error unmarshaling auth message: %v\n", err)
		s.sendError(conn, "Invalid authentication message format")
		authChan <- nil
		return
	}

	if msg.Action != "authenticate" {
		s.sendError(conn, "First message must be authentication")
		authChan <- nil
		return
	}

	token, ok := msg.Data.(string)
	if !ok {
		s.sendError(conn, "Invalid authentication token format")
		authChan <- nil
		return
	}

	user, err := s.verifyToken(token)
	if err != nil {
		log.Printf("Auth error: %v", err)
		s.sendError(conn, "Invalid authorization token")
		authChan <- nil
		return
	}

	s.conns.Store(conn, user.ID)
	log.Printf("User authenticated. ID: %s, Email: %s", user.ID, user.Email)

	conn.WriteJSON(types.Payload{
		Action: "authenticated",
		Data:   user.ID,
	})
	authChan <- user
}
