package realtime

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

const (
	maxMessageSize = 10 * 1024 * 1024
	pingInterval   = 10 * time.Second
	pingTimeout    = 60 * time.Second
)

type topics string

const (
	MonitorApplicationDeployment topics = "monitor_application_deployment"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SocketServer struct {
	conns            *sync.Map
	topicsMu         sync.RWMutex
	topics           map[string]map[*websocket.Conn]bool
	shutdown         chan struct{}
	deployController *deploy.DeployController
	db               *bun.DB
	ctx              context.Context
}

func NewSocketServer(deployController *deploy.DeployController, db *bun.DB, ctx context.Context) (*SocketServer, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	server := &SocketServer{
		conns:            &sync.Map{},
		shutdown:         make(chan struct{}),
		deployController: deployController,
		db:               db,
		ctx:              ctx,
		topics:           make(map[string]map[*websocket.Conn]bool),
	}

	return server, nil
}

func (s *SocketServer) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	fmt.Printf("New connection from client: %s\n", conn.RemoteAddr())
	conn.SetReadLimit(maxMessageSize)

	if token != "" {
		if len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		user, err := s.verifyToken(token)
		if err != nil {
			log.Printf("Auth error: %v", err)
			s.sendError(conn, "Invalid authorization token")
			conn.Close()
			return
		}

		log.Printf("User authenticated via URL token. ID: %s, Email: %s", user.ID, user.Email)
		s.conns.Store(conn, user.ID)
		defer s.handleDisconnect(conn)
		s.readLoop(conn, user)
		return
	}

	s.conns.Store(conn, "")
	defer s.handleDisconnect(conn)

	authTimer := time.NewTimer(30 * time.Second)
	authChan := make(chan *types.User)

	go func() {
		select {
		case user := <-authChan:
			if user != nil {
				s.readLoop(conn, user)
			}
		case <-authTimer.C:
			log.Printf("Authentication timeout for client: %s", conn.RemoteAddr())
			s.sendError(conn, "Authentication timeout")
			conn.Close()
		}
	}()

	s.waitForAuth(conn, authChan)
}

func (s *SocketServer) handleDisconnect(conn *websocket.Conn) {
	userID, _ := s.conns.Load(conn)
	s.conns.Delete(conn)

	s.topicsMu.Lock()
	for topic, connections := range s.topics {
		if _, exists := connections[conn]; exists {
			delete(connections, conn)
			if len(connections) == 0 {
				delete(s.topics, topic)
			}
		}
	}

	s.topicsMu.Unlock()

	conn.Close()
	fmt.Printf("Client disconnected: %s (User ID: %v)\n", conn.RemoteAddr(), userID)
}

func (s *SocketServer) Shutdown() {
	close(s.shutdown)
	s.conns.Range(func(conn, _ interface{}) bool {
		conn.(*websocket.Conn).Close()
		return true
	})
}

func (s *SocketServer) handlePing(conn *websocket.Conn) {
	userID, _ := s.conns.Load(conn)
	fmt.Printf("Received ping from %s (User ID: %v)\n", conn.RemoteAddr(), userID)
}

func (s *SocketServer) sendError(conn *websocket.Conn, message string) {
	conn.WriteJSON(types.Payload{
		Action: "error",
		Data:   message,
	})
}
