package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal/features/dashboard"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
	"github.com/raghavyuva/nixopus-api/internal/features/terminal"
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
	conns             *sync.Map
	topicsMu          sync.RWMutex
	topics            map[string]map[*websocket.Conn]bool
	shutdown          chan struct{}
	deployController  *deploy.DeployController
	db                *bun.DB
	ctx               context.Context
	postgres_listener PostgresListener
	terminalMutex     sync.RWMutex
	terminals         map[*websocket.Conn]*terminal.Terminal
	dashboardMonitors map[*websocket.Conn]*dashboard.DashboardMonitor
	dashboardMutex    sync.Mutex
}

// NewSocketServer initializes and returns a new instance of SocketServer.
// It sets up a PostgreSQL listener for application change notifications and
// starts a goroutine to handle these notifications.
//
// Parameters:
//   deployController - a pointer to an instance of DeployController used for handling deployment-related operations.
//   db - a pointer to a bun.DB object representing the database connection.

func NewSocketServer(deployController *deploy.DeployController, db *bun.DB, ctx context.Context) (*SocketServer, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pgListener := NewPostgresListener()

	server := &SocketServer{
		conns:             &sync.Map{},
		shutdown:          make(chan struct{}),
		deployController:  deployController,
		db:                db,
		ctx:               ctx,
		topics:            make(map[string]map[*websocket.Conn]bool),
		postgres_listener: *pgListener,
		terminals:         make(map[*websocket.Conn]*terminal.Terminal),
		dashboardMonitors: make(map[*websocket.Conn]*dashboard.DashboardMonitor),
	}

	notificationChan, err := pgListener.ListenToApplicationChanges(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to listen for PostgreSQL notifications: %w", err)
	}

	go server.handleNotifications(notificationChan)

	return server, nil
}

// handleNotifications processes incoming PostgreSQL notifications from the notification channel.
//
// This method listens on the provided channel for notifications related to application changes.
// Upon receiving a notification, it checks if the notification is from the "application_changes"
// channel. If it is, the method attempts to parse the JSON payload to extract the table, action,
// and ID details. If parsing is successful, it constructs a message containing the parsed data
// and broadcasts it to the appropriate topic using the BroadcastToTopic method.
//
// Parameters:
//
//	notificationChan - a channel that receives *PostgresNotification objects.
//
// Errors and any issues parsing the notification payload are logged, but do not stop the
// processing of further notifications.
func (s *SocketServer) handleNotifications(notificationChan <-chan *PostgresNotification) {
	for notification := range notificationChan {
		// fmt.Printf("Received notification on channel %s: %s\n",
		// 	notification.Channel, notification.Payload)

		if notification.Channel == "application_changes" {
			var parsedPayload struct {
				Table         string                 `json:"table"`
				Action        string                 `json:"action"`
				ApplicationID string                 `json:"application_id"`
				Data          map[string]interface{} `json:"data"`
			}

			if err := json.Unmarshal([]byte(notification.Payload), &parsedPayload); err != nil {
				log.Printf("Error parsing notification payload: %v", err)
				continue
			}

			resourceID := parsedPayload.ApplicationID

			messageData := map[string]interface{}{
				"table":          parsedPayload.Table,
				"action":         parsedPayload.Action,
				"application_id": parsedPayload.ApplicationID,
				"data":           parsedPayload.Data,
			}

			s.BroadcastToTopic(MonitorApplicationDeployment, resourceID, messageData)
		}
	}
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
