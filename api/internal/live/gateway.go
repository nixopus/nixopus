// Gateway provides real-time WebSocket connectivity for live development sessions, enabling developers to sync code changes instantly to cloud based development environments.
// this is the cloud part of syncing the local changes
package live

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	betterauth "github.com/raghavyuva/nixopus-api/internal/features/auth"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/mover"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	chunkSize             = int64(64 * 1024)
	fileCompletionWorkers = 8
	completionJobBuffer   = 128
)

type fileCompletionJob struct {
	content       []byte
	path          string
	checksum      string
	stagingPath   string
	appCtx        *ApplicationContext
	conn          *websocket.Conn
	applicationID uuid.UUID
}

type Gateway struct {
	stagingManager    *StagingManager
	buildFirstManager *BuildFirstManager
	websocketHandler  *WebSocketHandler
	manifestStore     *ManifestStore
	store             *shared_storage.Store
	logger            logger.Logger
	completionJobs    chan fileCompletionJob

	// sessionEnvStore: env vars from client (set-env file values, never the file itself)
	sessionEnvStore   map[string]map[string]string
	sessionEnvStoreMu sync.RWMutex
}

func NewGateway(stagingManager *StagingManager, taskService *tasks.TaskService, store *shared_storage.Store) *Gateway {
	logger := logger.NewLogger()
	gateway := &Gateway{
		stagingManager:  stagingManager,
		manifestStore:   NewManifestStore(),
		store:           store,
		logger:          logger,
		completionJobs:  make(chan fileCompletionJob, completionJobBuffer),
		sessionEnvStore: make(map[string]map[string]string),
	}
	gateway.buildFirstManager = NewBuildFirstManager(stagingManager, taskService, logger, func(appID uuid.UUID) map[string]string {
		return gateway.GetSessionEnv(appID)
	})
	gateway.websocketHandler = NewWebSocketHandler(gateway, logger)
	for i := 0; i < fileCompletionWorkers; i++ {
		go gateway.runFileCompletionWorker()
	}
	return gateway
}

// runFileCompletionWorker processes file write + ACK in parallel (non-blocking for read loop)
func (g *Gateway) runFileCompletionWorker() {
	for job := range g.completionJobs {
		ctx := context.WithValue(context.Background(), shared_types.OrganizationIDKey, job.appCtx.OrganizationID.String())
		if err := WriteContentToStaging(ctx, job.stagingPath, job.path, job.content, job.checksum); err != nil {
			g.logger.Log(logger.Error, "failed to write file to staging", fmt.Sprintf("path=%s err=%v", job.path, err))
			continue
		}
		g.manifestStore.Set(job.applicationID.String(), job.path, job.checksum)
		g.logger.Log(logger.Info, "file received and written", job.path)
		if g.buildFirstManager != nil {
			g.buildFirstManager.HandleFileWritten(ctx, job.appCtx, job.path, job.content)
		}
		if err := g.websocketHandler.sendAck(job.conn, job.path); err != nil {
			g.logger.Log(logger.Warning, "failed to send ACK", fmt.Sprintf("path=%s err=%v", job.path, err))
		}
	}
}

// BuildFirstManager returns the build-first manager for wiring callbacks (e.g. from TaskService).
func (g *Gateway) BuildFirstManager() *BuildFirstManager {
	return g.buildFirstManager
}

// GetSessionEnv returns env vars sent from client (set-env file values) for an application.
func (g *Gateway) GetSessionEnv(applicationID uuid.UUID) map[string]string {
	g.sessionEnvStoreMu.RLock()
	defer g.sessionEnvStoreMu.RUnlock()
	if env, ok := g.sessionEnvStore[applicationID.String()]; ok {
		return env
	}
	return nil
}

// HandleWebSocket delegates to the WebSocket handler
func (g *Gateway) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	g.websocketHandler.HandleWebSocket(w, r)
}

// VerifySession verifies the Better Auth session token and returns the user and organization ID
func (g *Gateway) VerifySession(ctx context.Context, tokenString string, originalRequest *http.Request) (*shared_types.User, string, error) {
	var req *http.Request

	// Prefer using the original request with actual cookies from the browser
	// WebSocket upgrade requests include cookies, which Better Auth needs
	if originalRequest != nil {
		// Clone the request to avoid modifying the original
		req = originalRequest.Clone(originalRequest.Context())
		// Also add the token as Authorization header as fallback
		if tokenString != "" {
			req.Header.Set("Authorization", "Bearer "+tokenString)
		}
	} else {
		// Fallback: create a mock request with the token
		req, _ = http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		// Also set as cookie if it's a cookie-based session
		req.AddCookie(&http.Cookie{
			Name:  "better-auth.session_token",
			Value: tokenString,
		})
	}

	// Verify Better Auth session - this is the source of truth for authentication and organization membership
	sessionResp, err := betterauth.VerifySession(req)
	if err != nil {
		return nil, "", fmt.Errorf("session verification failed: %v", err)
	}

	// Parse Better Auth user ID
	betterAuthUserID, err := uuid.Parse(sessionResp.User.ID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid user ID format: %v", err)
	}

	// Extract organization ID from session
	var orgID string
	if sessionResp.Session.ActiveOrganizationID != nil && *sessionResp.Session.ActiveOrganizationID != "" {
		orgID = *sessionResp.Session.ActiveOrganizationID
	} else {
		// Fallback to header if not in session
		orgID = originalRequest.Header.Get("X-Organization-Id")
	}

	// Create User object directly from Better Auth session response
	user := &shared_types.User{
		ID:            betterAuthUserID,
		Name:          sessionResp.User.Name,
		Email:         sessionResp.User.Email,
		EmailVerified: sessionResp.User.EmailVerified,
		CreatedAt:     time.Now(), // Better Auth doesn't provide this, use current time
		UpdatedAt:     time.Now(), // Better Auth doesn't provide this, use current time
	}

	// Compute backward compatibility fields
	user.ComputeCompatibilityFields()

	return user, orgID, nil
}

// handleMessage routes messages to appropriate handlers based on message type
func (g *Gateway) handleMessage(ctx context.Context, conn *websocket.Conn, appCtx *ApplicationContext, msg *mover.SyncMessage) error {
	switch msg.Type {
	case mover.MessageTypeFileChange:
		return g.handleFileChange(appCtx, msg)
	case mover.MessageTypeFileContent:
		return g.handleFileContent(ctx, conn, appCtx, msg)
	case mover.MessageTypeFileDelete:
		return g.handleFileDelete(ctx, appCtx, msg)
	case mover.MessageTypeEnvVars:
		return g.handleEnvVars(ctx, appCtx, msg)
	case mover.MessageTypePing:
		return g.websocketHandler.sendPong(conn)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (g *Gateway) unmarshalPayload(payload interface{}, target interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return nil
}

// validateFilePath validates that a file path is safe and doesn't escape the base_path
// Paths from the client are relative to base_path, so we validate:
// 1. No path traversal ("..")
// 2. No absolute paths
// 3. Path is clean and normalized
func (g *Gateway) validateFilePath(filePath string, basePath string) error {
	// Normalize to forward slashes for consistent checking
	normalizedPath := filepath.ToSlash(filePath)

	// Reject absolute paths
	if filepath.IsAbs(filePath) || strings.HasPrefix(normalizedPath, "/") {
		return fmt.Errorf("absolute paths are not allowed: %s", filePath)
	}

	// Reject paths that try to escape (path traversal)
	if strings.Contains(normalizedPath, "..") {
		return fmt.Errorf("path traversal detected: %s", filePath)
	}

	// Clean the path to resolve any remaining issues
	cleanPath := filepath.Clean(normalizedPath)

	// If base_path is "/" or empty, all relative paths are valid
	// Otherwise, paths should not try to escape the base_path context
	// Since paths are already relative to base_path from the client,
	// we just need to ensure they don't contain traversal
	if cleanPath != normalizedPath && strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path after cleaning: %s", filePath)
	}

	return nil
}

func (g *Gateway) handleEnvVars(ctx context.Context, appCtx *ApplicationContext, msg *mover.SyncMessage) error {
	var payload mover.EnvVarsPayload
	if err := g.unmarshalPayload(msg.Payload, &payload); err != nil {
		return err
	}
	if len(payload.Vars) == 0 {
		return nil
	}
	g.sessionEnvStoreMu.Lock()
	g.sessionEnvStore[appCtx.ApplicationID.String()] = payload.Vars
	g.sessionEnvStoreMu.Unlock()
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, appCtx.OrganizationID.String())
	if err := tasks.UpdateLiveDevServiceEnv(orgCtx, appCtx.ApplicationID, payload.Vars); err != nil {
		g.logger.Log(logger.Warning, "failed to update service env", fmt.Sprintf("app=%s err=%v", appCtx.ApplicationID, err))
	} else {
		g.logger.Log(logger.Info, "env vars updated, service rolling new tasks", appCtx.ApplicationID.String())
	}
	return nil
}

func (g *Gateway) handleFileChange(appCtx *ApplicationContext, msg *mover.SyncMessage) error {
	var fileChange mover.FileChange
	if err := g.unmarshalPayload(msg.Payload, &fileChange); err != nil {
		return err
	}

	// Validate file path is within base_path
	// Paths from client are relative to base_path, so we validate they don't escape
	if err := g.validateFilePath(fileChange.Path, appCtx.BasePath); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	totalChunks := int((fileChange.Size + chunkSize - 1) / chunkSize)
	if totalChunks == 0 {
		totalChunks = 1
	}
	g.stagingManager.GetFileReceiver(appCtx.ApplicationID, fileChange.Path, totalChunks, fileChange.Checksum, appCtx.StagingPath)

	g.logger.Log(logger.Info, "file change received", fmt.Sprintf("path=%s op=%s size=%d chunks=%d", fileChange.Path, fileChange.Operation, fileChange.Size, totalChunks))
	return nil
}

func (g *Gateway) handleFileContent(ctx context.Context, conn *websocket.Conn, appCtx *ApplicationContext, msg *mover.SyncMessage) error {
	var fileContent mover.FileContent
	if err := g.unmarshalPayload(msg.Payload, &fileContent); err != nil {
		return err
	}

	// Validate file path is within base_path
	// Paths from client are relative to base_path, so we validate they don't escape
	if err := g.validateFilePath(fileContent.Path, appCtx.BasePath); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	receiver := g.stagingManager.GetFileReceiver(appCtx.ApplicationID, fileContent.Path, fileContent.TotalChunks, fileContent.Checksum, appCtx.StagingPath)

	// Update metadata first to ensure consistent state
	receiver.UpdateMetadata(fileContent.TotalChunks, fileContent.Checksum)

	// Add chunk with validation
	if err := receiver.AddChunk(fileContent.ChunkIndex, fileContent.Data); err != nil {
		return fmt.Errorf("failed to add chunk: %w", err)
	}

	if !receiver.IsComplete() {
		return nil
	}

	content, err := receiver.Reassemble()
	if err != nil {
		return fmt.Errorf("failed to reassemble file: %w", err)
	}

	g.stagingManager.RemoveFileReceiver(appCtx.ApplicationID, fileContent.Path)

	// Offload write + ACK to worker pool - read loop continues immediately
	job := fileCompletionJob{
		content:       content,
		path:          fileContent.Path,
		checksum:      fileContent.Checksum,
		stagingPath:   appCtx.StagingPath,
		appCtx:        appCtx,
		conn:          conn,
		applicationID: appCtx.ApplicationID,
	}
	select {
	case g.completionJobs <- job:
		// Queued - worker will handle it
	default:
		// Queue full - fall back to synchronous write to avoid dropping
		if err := WriteContentToStaging(ctx, job.stagingPath, job.path, job.content, job.checksum); err != nil {
			return fmt.Errorf("failed to write file to staging: %w", err)
		}
		g.manifestStore.Set(job.applicationID.String(), job.path, job.checksum)
		g.logger.Log(logger.Info, "file received and written", job.path)
		if g.buildFirstManager != nil {
			g.buildFirstManager.HandleFileWritten(ctx, job.appCtx, job.path, job.content)
		}
		if err := g.websocketHandler.sendAck(conn, fileContent.Path); err != nil {
			g.logger.Log(logger.Warning, "failed to send ACK", fmt.Sprintf("path=%s err=%v", fileContent.Path, err))
		}
	}

	return nil
}

func (g *Gateway) handleFileDelete(ctx context.Context, appCtx *ApplicationContext, msg *mover.SyncMessage) error {
	var fileChange mover.FileChange
	if err := g.unmarshalPayload(msg.Payload, &fileChange); err != nil {
		return err
	}

	// Validate file path is within base_path
	// Paths from client are relative to base_path, so we validate they don't escape
	if err := g.validateFilePath(fileChange.Path, appCtx.BasePath); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	if err := DeleteFileFromStaging(ctx, appCtx.StagingPath, fileChange.Path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	g.manifestStore.Remove(appCtx.ApplicationID.String(), fileChange.Path)

	g.logger.Log(logger.Info, "file deleted", fileChange.Path)
	if g.buildFirstManager != nil {
		g.buildFirstManager.HandleFileDeleted(ctx, appCtx, fileChange.Path)
	}
	return nil
}
