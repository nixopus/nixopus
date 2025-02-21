package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	maxMessageSize = 10 * 1024 * 1024
	pingInterval   = 10 * time.Second
	pingTimeout    = 60 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SocketServer struct {
	conns    *sync.Map
	shutdown chan struct{}
}

func NewSocketServer() (*SocketServer, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	server := &SocketServer{
		conns:    &sync.Map{},
		shutdown: make(chan struct{}),
	}

	return server, nil
}

func (s *SocketServer) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	fmt.Println("New incoming connection from client:", conn.RemoteAddr())

	conn.SetReadLimit(maxMessageSize)
	// conn.SetReadDeadline(time.Now().Add(pongWait))
	// conn.SetPongHandler(func(string) error {
	// 	conn.SetReadDeadline(time.Now().Add(pongWait))
	// 	return nil
	// })

	s.conns.Store(conn, "")
	defer s.handleDisconnect(conn)
	s.readLoop(conn)
}

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
		case "ping":
			s.handleMessage(conn, msg)
		default:
			s.sendError(conn, "Unknown message action")
		}
	}
}

func (s *SocketServer) handleDisconnect(conn *websocket.Conn) {
	conn.Close()
}

func (s *SocketServer) Shutdown() {
	close(s.shutdown)
	s.conns.Range(func(conn, _ interface{}) bool {
		conn.(*websocket.Conn).Close()
		return true
	})
}

func (s *SocketServer) handleMessage(conn *websocket.Conn, msg types.Payload) {
	fmt.Printf("Received message from %s: %v with payload %v\n", conn.RemoteAddr(), msg.Action, msg.Data)
	switch msg.Action {
	case "ping":
		s.handlePing(conn)
	}
}

func (s *SocketServer) handlePing(conn *websocket.Conn) {
	fmt.Printf("Received ping from %s\n", conn.RemoteAddr())
}

func (s *SocketServer) sendError(conn *websocket.Conn, message string) {
	conn.WriteJSON(types.Payload{
		Action: "error",
		Data:   message,
	})
}
