package live

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/mover"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

var upgrader websocket.Upgrader
var upgraderOnce sync.Once

func getUpgrader() *websocket.Upgrader {
	upgraderOnce.Do(func() {
		upgrader = websocket.Upgrader{
			ReadBufferSize:  readBufferSize(),
			WriteBufferSize: writeBufferSize(),
			CheckOrigin:     checkOriginFunc(),
		}
	})
	return &upgrader
}

// WebSocketHandler manages WebSocket connections and message processing.
// Uses per-connection write mutexes to avoid serializing writes across different connections.
type WebSocketHandler struct {
	gateway     *Gateway
	logger      logger.Logger
	connWriteMu sync.Map // map[*websocket.Conn]*sync.Mutex — per-connection write serialization
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(gateway *Gateway, logger logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		gateway: gateway,
		logger:  logger,
	}
}

// HandleWebSocket handles incoming WebSocket connections for live development sessions.
// It authenticates the client, validates application ownership, and establishes a real-time connection for file synchronization.
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	applicationID, err := h.extractApplicationID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	user, orgID, err := h.gateway.VerifySession(ctx, token, r)
	if err != nil {
		h.logger.Log(logger.Error, "invalid session", err.Error())
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}

	// Parse organization ID
	organizationID, err := uuid.Parse(orgID)
	if err != nil {
		h.logger.Log(logger.Error, "invalid organization ID", err.Error())
		http.Error(w, "Invalid organization", http.StatusUnauthorized)
		return
	}

	// Set organization ID in context for downstream operations (SSH manager, etc.)
	ctx = context.WithValue(ctx, types.OrganizationIDKey, orgID)

	appCtx, err := h.gateway.getApplicationContext(ctx, r, token, applicationID, user.ID, organizationID)
	if err != nil {
		h.logger.Log(logger.Error, "failed to get application context", fmt.Sprintf("application_id=%s err=%v", applicationID, err))
		http.Error(w, fmt.Sprintf("Application not found or access denied: %v", err), http.StatusNotFound)
		return
	}

	conn, err := getUpgrader().Upgrade(w, r, nil)
	if err != nil {
		h.logger.Log(logger.Error, "websocket upgrade failed", err.Error())
		return
	}
	defer conn.Close()

	h.logger.Log(logger.Info, "websocket connection established", fmt.Sprintf("application_id=%s", applicationID))

	// Track this connection so pipeline progress can be sent to the client
	h.gateway.registerConn(applicationID, conn, h)
	defer func() {
		h.cleanupConnMutex(conn)
		h.gateway.unregisterConn(applicationID)
	}()

	// Send manifest immediately so client can skip already-synced files
	if err := h.sendManifest(ctx, conn, appCtx.ApplicationID); err != nil {
		h.logger.Log(logger.Warning, "failed to send manifest", err.Error())
	}

	h.processMessages(ctx, conn, appCtx)
}

// extractApplicationID extracts the application ID from the WebSocket path
func (h *WebSocketHandler) extractApplicationID(path string) (uuid.UUID, error) {
	prefix := "/ws/live/"
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, fmt.Errorf("invalid path")
	}
	applicationIDStr := path[len(prefix):]
	if applicationIDStr == "" {
		return uuid.Nil, fmt.Errorf("missing application ID")
	}
	return uuid.Parse(applicationIDStr)
}

// processMessages continuously reads and processes messages from the WebSocket connection.
// It handles file changes, content chunks, deletions, and ping/pong messages for maintaining the connection.
func (h *WebSocketHandler) processMessages(ctx context.Context, conn *websocket.Conn, appCtx *ApplicationContext) {
	h.setupConnectionHandlers(conn)
	conn.SetReadDeadline(time.Now().Add(readDeadline()))

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			h.handleReadError(err, appCtx.ApplicationID)
			break
		}

		conn.SetReadDeadline(time.Now().Add(readDeadline()))

		if messageType == websocket.PingMessage || messageType == websocket.PongMessage {
			continue
		}

		var msg recvMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			h.logger.Log(logger.Error, "failed to parse message", err.Error())
			h.sendError(conn, "invalid_message", "Failed to parse message")
			continue
		}

		h.processMessage(ctx, conn, appCtx, &msg)
	}

	h.logger.Log(logger.Info, "websocket connection closed", fmt.Sprintf("application_id=%s", appCtx.ApplicationID))
}

// setupConnectionHandlers configures ping/pong handlers for connection keepalive
func (h *WebSocketHandler) setupConnectionHandlers(conn *websocket.Conn) {
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(readDeadline()))
		return nil
	})

	conn.SetPingHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(readDeadline()))
		conn.SetWriteDeadline(time.Now().Add(writeDeadline()))
		mu := h.getConnMutex(conn)
		mu.Lock()
		defer mu.Unlock()
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(writeDeadline()))
	})
}

// handleReadError handles WebSocket read errors and logs them appropriately
func (h *WebSocketHandler) handleReadError(err error, applicationID uuid.UUID) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		h.logger.Log(logger.Error, "websocket error", err.Error())
	} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		h.logger.Log(logger.Warning, "websocket read deadline expired", applicationID.String())
	} else {
		h.logger.Log(logger.Error, "websocket read error", err.Error())
	}
}

// recvMessage is the receive-only message shape. Payload as json.RawMessage avoids
// double encode: we unmarshal once, then unmarshal Payload directly into typed structs.
type recvMessage struct {
	Type      mover.MessageType `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	Payload   json.RawMessage   `json:"payload"`
}

// processMessage processes a single message and delegates to the gateway for handling
func (h *WebSocketHandler) processMessage(ctx context.Context, conn *websocket.Conn, appCtx *ApplicationContext, msg *recvMessage) {
	if err := h.gateway.handleMessage(ctx, conn, appCtx, msg); err != nil {
		h.logger.Log(logger.Error, "failed to handle message", err.Error())
		if sendErr := h.sendError(conn, "processing_error", err.Error()); sendErr != nil {
			h.logger.Log(logger.Error, "failed to send error message", sendErr.Error())
		}
	}
}

// sendError sends an error message to the client
func (h *WebSocketHandler) sendError(conn *websocket.Conn, code, message string) error {
	return h.sendMessage(conn, mover.SyncMessage{
		Type:      mover.MessageTypeError,
		Timestamp: time.Now(),
		Payload:   mover.ErrorPayload{Code: code, Message: message},
	})
}

// sendAck sends an acknowledgment message for a received file
func (h *WebSocketHandler) sendAck(conn *websocket.Conn, filePath string) error {
	return h.sendMessage(conn, mover.SyncMessage{
		Type:      mover.MessageTypeAck,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"file_path": filePath, "status": "received"},
	})
}

// sendManifest sends the server's file manifest to the client for incremental sync.
// Includes root_hash so client can skip full diff when roots match.
// Always loads from DB (single source of truth).
func (h *WebSocketHandler) sendManifest(ctx context.Context, conn *websocket.Conn, applicationID uuid.UUID) error {
	paths, _, err := LoadManifest(ctx, h.gateway.store, applicationID)
	if err != nil {
		return err
	}
	if paths == nil {
		paths = make(map[string]string)
	}
	tree := mover.BuildFromPaths(paths)
	return h.sendMessage(conn, mover.SyncMessage{
		Type:      mover.MessageTypeManifest,
		Timestamp: time.Now(),
		Payload:   mover.ManifestPayload{Paths: paths, RootHash: tree.RootHash, Version: 1},
	})
}

// sendPong sends a pong message in response to a ping.
// Uses sendMessage for proper per-connection write serialization (avoids race with other writers).
func (h *WebSocketHandler) sendPong(conn *websocket.Conn) error {
	return h.sendMessage(conn, mover.SyncMessage{
		Type:      mover.MessageTypePong,
		Timestamp: time.Now(),
	})
}

// sendMessage sends a message to the WebSocket connection with per-connection write mutex.
// Gorilla websocket allows one concurrent writer per connection; each conn has its own mutex
// so different connections can write concurrently without blocking each other.
func (h *WebSocketHandler) sendMessage(conn *websocket.Conn, msg mover.SyncMessage) error {
	mu := h.getConnMutex(conn)
	mu.Lock()
	defer mu.Unlock()
	conn.SetWriteDeadline(time.Now().Add(writeDeadline()))
	return conn.WriteJSON(msg)
}

// getConnMutex returns (or creates) the write mutex for a connection.
func (h *WebSocketHandler) getConnMutex(conn *websocket.Conn) *sync.Mutex {
	if v, ok := h.connWriteMu.Load(conn); ok {
		return v.(*sync.Mutex)
	}
	mu := &sync.Mutex{}
	if v, loaded := h.connWriteMu.LoadOrStore(conn, mu); loaded {
		return v.(*sync.Mutex)
	}
	return mu
}

// cleanupConnMutex removes the mutex for a closed connection to avoid leaks.
// Must be called when the connection is closed (e.g. in defer).
func (h *WebSocketHandler) cleanupConnMutex(conn *websocket.Conn) {
	h.connWriteMu.Delete(conn)
}

func (g *Gateway) getApplicationContext(ctx context.Context, r *http.Request, token string, applicationID, userID, organizationID uuid.UUID) (*ApplicationContext, error) {
	// Get application
	deployStorage := storage.DeployStorage{DB: g.store.DB, Ctx: ctx}
	application, err := deployStorage.GetApplicationById(applicationID.String(), organizationID)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	// Validate ownership
	if application.UserID != userID || application.OrganizationID != organizationID {
		return nil, fmt.Errorf("application does not belong to user/organization")
	}

	// Get staging path using existing GetClonePath function
	stagingPath, err := g.stagingManager.GetStagingPath(ctx, applicationID, userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get staging path: %w", err)
	}

	// Generate domain name based on application ID: {first-8-chars}.{deploy_domain}
	domain := fmt.Sprintf("%s.%s", applicationID.String()[:8], config.GetDeployDomain())

	// Parse environment variables from application
	envVars := tasks.GetMapFromString(application.EnvironmentVariables)

	// Get base_path from application (default to "/" if empty)
	basePath := application.BasePath
	if basePath == "" {
		basePath = "/"
	}

	repoSource := g.stagingManager.GetRepositorySource(ctx, &application)

	authCookie := ""
	if r != nil {
		authCookie = r.Header.Get("Cookie")
	}

	return &ApplicationContext{
		ApplicationID:        applicationID,
		UserID:               userID,
		OrganizationID:       organizationID,
		StagingPath:          stagingPath,
		RepositorySource:     repoSource,
		BasePath:             basePath,
		Environment:          application.Environment,
		Domain:               domain,
		Config:               make(map[string]interface{}),
		EnvironmentVariables: envVars,
		AuthToken:            token,
		AuthCookie:           authCookie,
	}, nil
}
