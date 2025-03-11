package realtime

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// SubscribeToTopic adds a connection to the specified topic.
//
// The function takes a topic string and a connection as parameters and
// stores the connection in the topic map. It is safe to call this function
// concurrently from multiple goroutines.
func (s *SocketServer) SubscribeToTopic(topic topics, resourceID string, conn *websocket.Conn) {
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()
	var topicKey string

	if resourceID == "" {
		topicKey = string(topic)
	} else {
		topicKey = fmt.Sprintf("%s:%s", string(topic), resourceID)
	}

	if _, exists := s.topics[topicKey]; !exists {
		s.topics[topicKey] = make(map[*websocket.Conn]bool)
	}
	s.topics[topicKey][conn] = true

	conn.WriteJSON(types.Payload{
		Action: "subscribed",
		Topic:  string(topicKey),
		Data:   nil,
	})

	log.Printf("Client %s subscribed to topic %s", conn.RemoteAddr(), topicKey)
}

// UnsubscribeFromTopic removes a connection from the specified topic.
//
// The function takes a topic string and a connection as parameters and
// removes the connection from the topic map. It is safe to call this function
// concurrently from multiple goroutines.
func (s *SocketServer) UnsubscribeFromTopic(topic topics, resourceID string, conn *websocket.Conn) {
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()
	var topicKey string

	if resourceID == "" {
		topicKey = string(topic)
	} else {
		topicKey = fmt.Sprintf("%s:%s", string(topic), resourceID)
	}

	if connections, exists := s.topics[topicKey]; exists {
		delete(connections, conn)

		if len(connections) == 0 {
			delete(s.topics, topicKey)
		}

		conn.WriteJSON(types.Payload{
			Action: "unsubscribed",
			Topic:  string(topicKey),
			Data:   nil,
		})

		log.Printf("Client %s unsubscribed from topic %s", conn.RemoteAddr(), topicKey)
	}
}

// BroadcastToTopic sends a message to all connections subscribed to the specified topic.
//
// The function takes a topic string and a payload as parameters and
// sends the payload to all connections subscribed to the topic.
func (s *SocketServer) BroadcastToTopic(topic topics, resourceID string, payload interface{}) {
	s.topicsMu.RLock()
	defer s.topicsMu.RUnlock()
	var topicKey string

	if resourceID == "" {
		topicKey = string(topic)
	} else {
		topicKey = fmt.Sprintf("%s:%s", string(topic), resourceID)
	}

	if connections, exists := s.topics[topicKey]; exists {
		for conn := range connections {
			err := conn.WriteJSON(types.Payload{
				Action: "message",
				Topic:  string(topicKey),
				Data:   payload,
			})

			if err != nil {
				log.Printf("Error broadcasting to client %s: %v", conn.RemoteAddr(), err)
				go func(c *websocket.Conn) {
					s.UnsubscribeFromTopic(topic, resourceID, c)
				}(conn)
			}
		}
		log.Printf("Broadcast message to %d clients on topic %s", len(connections), topicKey)
	}
}
