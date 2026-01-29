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
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/mover"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

const (
	readDeadline  = 5 * time.Minute
	writeDeadline = 10 * time.Second
)

// WebSocketHandler manages WebSocket connections and message processing
type WebSocketHandler struct {
	gateway *Gateway
	logger  logger.Logger
	writeMu sync.Mutex
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
	apiKey, err := h.gateway.verifyAPIKey(ctx, token)
	if err != nil {
		h.logger.Log(logger.Error, "invalid API key", err.Error())
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}

	// Set organization ID in context for downstream operations (SSH manager, etc.)
	ctx = context.WithValue(ctx, types.OrganizationIDKey, apiKey.OrganizationID.String())

	// Get application and validate ownership
	appCtx, err := h.gateway.getApplicationContext(ctx, applicationID, apiKey.UserID, apiKey.OrganizationID)
	if err != nil {
		h.logger.Log(logger.Error, "failed to get application context", fmt.Sprintf("application_id=%s err=%v", applicationID, err))
		http.Error(w, fmt.Sprintf("Application not found or access denied: %v", err), http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Log(logger.Error, "websocket upgrade failed", err.Error())
		return
	}
	defer conn.Close()

	h.logger.Log(logger.Info, "websocket connection established", fmt.Sprintf("application_id=%s", applicationID))
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
	conn.SetReadDeadline(time.Now().Add(readDeadline))

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			h.handleReadError(err, appCtx.ApplicationID)
			break
		}

		conn.SetReadDeadline(time.Now().Add(readDeadline))

		if messageType == websocket.PingMessage || messageType == websocket.PongMessage {
			continue
		}

		var msg mover.SyncMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			h.logger.Log(logger.Error, "failed to parse message", err.Error())
			h.sendError(conn, "invalid_message", "Failed to parse message")
			continue
		}

		h.processMessage(ctx, conn, appCtx, msg)
	}

	h.logger.Log(logger.Info, "websocket connection closed", fmt.Sprintf("application_id=%s", appCtx.ApplicationID))
}

// setupConnectionHandlers configures ping/pong handlers for connection keepalive
func (h *WebSocketHandler) setupConnectionHandlers(conn *websocket.Conn) {
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	conn.SetPingHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(readDeadline))
		conn.SetWriteDeadline(time.Now().Add(writeDeadline))
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(writeDeadline))
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

// processMessage processes a single message and delegates to the gateway for handling
func (h *WebSocketHandler) processMessage(ctx context.Context, conn *websocket.Conn, appCtx *ApplicationContext, msg mover.SyncMessage) {
	if err := h.gateway.handleMessage(ctx, conn, appCtx, &msg); err != nil {
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

// sendPong sends a pong message in response to a ping
func (h *WebSocketHandler) sendPong(conn *websocket.Conn) error {
	return conn.WriteJSON(mover.SyncMessage{
		Type:      mover.MessageTypePong,
		Timestamp: time.Now(),
	})
}

// sendMessage sends a message to the WebSocket connection with write mutex protection
func (h *WebSocketHandler) sendMessage(conn *websocket.Conn, msg mover.SyncMessage) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	conn.SetWriteDeadline(time.Now().Add(writeDeadline))
	return conn.WriteJSON(msg)
}

// getApplicationContext gets application information and staging path
func (g *Gateway) getApplicationContext(ctx context.Context, applicationID, userID, organizationID uuid.UUID) (*ApplicationContext, error) {
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

	// Generate domain name based on application ID: {first-8-chars}.nixopus.com
	domain := fmt.Sprintf("%s.nixopus.com", applicationID.String()[:8])

	// Parse environment variables from application
	envVars := tasks.GetMapFromString(application.EnvironmentVariables)

	// Get base_path from application (default to "/" if empty)
	basePath := application.BasePath
	if basePath == "" {
		basePath = "/"
	}

	return &ApplicationContext{
		ApplicationID:        applicationID,
		UserID:               userID,
		OrganizationID:       organizationID,
		StagingPath:          stagingPath,
		BasePath:             basePath,
		Environment:          application.Environment,
		Domain:               domain,
		Config:               make(map[string]interface{}),
		EnvironmentVariables: envVars,
	}, nil
}
