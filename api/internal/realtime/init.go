package realtime

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/nixopus/nixopus/api/internal/features/dashboard"
	deploy "github.com/nixopus/nixopus/api/internal/features/deploy/controller"
	"github.com/nixopus/nixopus/api/internal/features/deploy/realtime"
	"github.com/nixopus/nixopus/api/internal/features/terminal"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

const (
	maxMessageSize = 10 * 1024 * 1024
	pingInterval   = 10 * time.Second
	pingTimeout    = 60 * time.Second
)

type Topics string

const (
	MonitorApplicationDeployment Topics = "monitor_application_deployment"
	MonitorHealthCheck           Topics = "monitor_health_check"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// LiveDevNotificationHandler is called for live_dev_logs and live_dev_status
// PostgreSQL notifications. The Gateway implements this to forward build logs
// and status changes to the appropriate WebSocket client.
type LiveDevNotificationHandler func(channel, payload string)

type SocketServer struct {
	conns               *sync.Map // conn -> userID
	orgIDs              *sync.Map // conn -> organizationID
	connWriteMu         sync.Map  // conn -> *sync.Mutex (per-connection write serialization)
	topicsMu            sync.RWMutex
	topics              map[string]map[*websocket.Conn]bool
	shutdown            chan struct{}
	deployController    *deploy.DeployController
	db                  *bun.DB
	ctx                 context.Context
	postgres_listener   PostgresListener
	terminalMutex       sync.RWMutex
	terminals           map[*websocket.Conn]map[string]*terminal.Terminal // conn -> terminalId -> terminal session for handling multiple terminal sessions per connection
	dashboardMonitors   map[*websocket.Conn]*dashboard.DashboardMonitor
	dashboardMutex      sync.Mutex
	applicationMonitors map[*websocket.Conn]*realtime.ApplicationMonitor
	applicationMutex    sync.Mutex

	liveDevHandler   LiveDevNotificationHandler
	liveDevHandlerMu sync.RWMutex
}

// NewSocketServer initializes and returns a new instance of SocketServer.
func NewSocketServer(deployController *deploy.DeployController, db *bun.DB, ctx context.Context) (*SocketServer, error) {
	// Load .env file if it exists (optional when using secret manager)
	_ = godotenv.Load()

	pgListener := NewPostgresListener()

	server := &SocketServer{
		conns:               &sync.Map{},
		orgIDs:              &sync.Map{},
		shutdown:            make(chan struct{}),
		deployController:    deployController,
		db:                  db,
		ctx:                 ctx,
		topics:              make(map[string]map[*websocket.Conn]bool),
		postgres_listener:   *pgListener,
		terminals:           make(map[*websocket.Conn]map[string]*terminal.Terminal),
		dashboardMonitors:   make(map[*websocket.Conn]*dashboard.DashboardMonitor),
		applicationMonitors: make(map[*websocket.Conn]*realtime.ApplicationMonitor),
	}
	err := StartListeningAndNotify(&server.postgres_listener, ctx, server)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// HandleHTTP handles incoming HTTP connections and upgrades them to WebSocket connections.
// It verifies the authorization token and authenticates the user.
//
// Parameters:
//
//	w - the http.ResponseWriter for the response.
//	r - the http.Request from the client.
//
// Returns:
//   - nil if the connection is successfully upgraded.
//   - an error if the connection fails to upgrade or if the token is invalid.
func (s *SocketServer) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusUnauthorized)
		return
	}

	orgID := r.URL.Query().Get("organization-id")
	if orgID != "" {
		r.Header.Set("X-Organization-Id", orgID)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	conn.SetReadLimit(maxMessageSize)

	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	user, orgID, err := s.verifyToken(token, r)
	if err != nil || user == nil {
		s.sendError(conn, "Invalid authorization token")
		conn.Close()
		return
	}

	s.conns.Store(conn, user.ID)
	if orgID != "" {
		s.orgIDs.Store(conn, orgID)
	}
	defer s.handleDisconnect(conn)

	s.readLoop(conn)
}

// handleDisconnect handles the disconnection of a client.
// it closes all the monitors and deletes the connection from the map.
// Parameters:
//
//	conn - the *websocket.Conn representing the client connection.
//
// Returns:
//   - nil
func (s *SocketServer) handleDisconnect(conn *websocket.Conn) {
	fmt.Println("[ws] handleDisconnect: client disconnected, cleaning up")
	s.conns.Delete(conn)
	s.orgIDs.Delete(conn)

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

	s.terminalMutex.Lock()
	if terminalSessions, exists := s.terminals[conn]; exists {
		fmt.Printf("[ws] handleDisconnect: closing %d terminal session(s)\n", len(terminalSessions))
		for id, terminalSession := range terminalSessions {
			fmt.Printf("[ws] handleDisconnect: closing terminal %s\n", id)
			terminalSession.Close()
		}
		delete(s.terminals, conn)
	}
	s.terminalMutex.Unlock()

	s.dashboardMutex.Lock()
	if monitor, exists := s.dashboardMonitors[conn]; exists {
		monitor.Stop()
		delete(s.dashboardMonitors, conn)
	}
	s.dashboardMutex.Unlock()

	s.applicationMutex.Lock()
	if monitor, exists := s.applicationMonitors[conn]; exists {
		monitor.Stop()
		delete(s.applicationMonitors, conn)
	}
	s.applicationMutex.Unlock()

	conn.Close()
	s.connWriteMu.Delete(conn)
}

// SetLiveDevHandler registers a handler for live_dev_logs and live_dev_status
// notifications from PostgreSQL. The live Gateway calls this during initialization.
func (s *SocketServer) SetLiveDevHandler(handler LiveDevNotificationHandler) {
	s.liveDevHandlerMu.Lock()
	s.liveDevHandler = handler
	s.liveDevHandlerMu.Unlock()
}

func (s *SocketServer) Shutdown() {
	close(s.shutdown)
	s.conns.Range(func(conn, _ interface{}) bool {
		conn.(*websocket.Conn).Close()
		return true
	})
}

func (s *SocketServer) handlePing(conn *websocket.Conn) {
	_, _ = s.conns.Load(conn)
}

// getConnWriteMu returns the per-connection write mutex, creating one if needed.
func (s *SocketServer) getConnWriteMu(conn *websocket.Conn) *sync.Mutex {
	if mu, ok := s.connWriteMu.Load(conn); ok {
		return mu.(*sync.Mutex)
	}
	mu := &sync.Mutex{}
	actual, _ := s.connWriteMu.LoadOrStore(conn, mu)
	return actual.(*sync.Mutex)
}

// writeJSON serializes all writes to a connection through its per-connection mutex.
func (s *SocketServer) writeJSON(conn *websocket.Conn, v interface{}) error {
	mu := s.getConnWriteMu(conn)
	mu.Lock()
	defer mu.Unlock()
	return conn.WriteJSON(v)
}

func (s *SocketServer) sendError(conn *websocket.Conn, message string) {
	s.writeJSON(conn, types.Payload{
		Action: "error",
		Data:   message,
	})
}
