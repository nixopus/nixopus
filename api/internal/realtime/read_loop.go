package realtime

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *SocketServer) readLoop(conn *websocket.Conn, user *types.User) {
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

		default:
			s.sendError(conn, "Unknown message action")
		}
	}
}
