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
	"sync/atomic"
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

type fileCompletionJob struct {
	content       []byte
	path          string
	checksum      string
	stagingPath   string
	appCtx        *ApplicationContext
	conn          *websocket.Conn
	applicationID uuid.UUID
}

var jobPool = sync.Pool{
	New: func() interface{} { return &fileCompletionJob{} },
}

func getJob() *fileCompletionJob { return jobPool.Get().(*fileCompletionJob) }
func putJob(job *fileCompletionJob) {
	job.content, job.appCtx, job.conn = nil, nil, nil
	job.path, job.checksum, job.stagingPath = "", "", ""
	job.applicationID = uuid.Nil
	jobPool.Put(job)
}

type Gateway struct {
	stagingManager    *StagingManager
	buildFirstManager *BuildFirstManager
	websocketHandler  *WebSocketHandler
	store             *shared_storage.Store
	logger            logger.Logger
	completionJobs    chan *fileCompletionJob
	workerWg          sync.WaitGroup
	shuttingDown      atomic.Bool
	shutdownOnce      sync.Once

	// sessionEnvStore: env vars from client (set-env file values, never the file itself)
	sessionEnvStore   map[string]map[string]string
	sessionEnvStoreMu sync.RWMutex

	// activeConns tracks the WebSocket connection per application for sending pipeline progress
	activeConnsMu sync.RWMutex
	activeConns   map[uuid.UUID]*activeConn

	// pendingCompletions tracks in-flight file writes per app; build must wait for 0 before starting
	pendingCompletions   map[string]*atomic.Int64
	pendingCompletionsMu sync.RWMutex
}

type activeConn struct {
	conn    *websocket.Conn
	handler *WebSocketHandler
}

func NewGateway(stagingManager *StagingManager, taskService *tasks.TaskService, store *shared_storage.Store) *Gateway {
	logger := logger.NewLogger()
	gateway := &Gateway{
		stagingManager:     stagingManager,
		store:              store,
		logger:             logger,
		completionJobs:     make(chan *fileCompletionJob, completionBuffer()),
		sessionEnvStore:    make(map[string]map[string]string),
		activeConns:        make(map[uuid.UUID]*activeConn),
		pendingCompletions: make(map[string]*atomic.Int64),
	}
	gateway.buildFirstManager = NewBuildFirstManager(stagingManager, taskService, logger, func(appID uuid.UUID) map[string]string {
		return gateway.GetSessionEnv(appID)
	})
	gateway.buildFirstManager.SetPipelineProgressFunc(func(appID uuid.UUID, stageId, message string) {
		gateway.sendPipelineProgress(appID, stageId, message)
	})
	gateway.buildFirstManager.SetBuildStatusFunc(func(appID uuid.UUID, phase, message, errMsg string) {
		gateway.sendBuildStatus(appID, phase, message, errMsg)
	})
	gateway.buildFirstManager.SetCodebaseIndexedFunc(func(appCtx *ApplicationContext) {
		_ = gateway.sendCodebaseIndexed(appCtx)
	})
	gateway.websocketHandler = NewWebSocketHandler(gateway, logger)
	for i := 0; i < completionWorkers(); i++ {
		gateway.workerWg.Add(1)
		go func() {
			defer gateway.workerWg.Done()
			gateway.runFileCompletionWorker()
		}()
	}
	return gateway
}

// Shutdown gracefully stops the Gateway: signals completion workers to drain and waits for them.
// After Shutdown, new file completions will run inline (completeFileSync) instead of queuing.
// Safe to call multiple times; only the first call performs shutdown.
func (g *Gateway) Shutdown() {
	g.shutdownOnce.Do(func() {
		g.shuttingDown.Store(true)
		close(g.completionJobs)
		g.workerWg.Wait()
	})
}

// registerConn tracks an active WebSocket connection for an application.
func (g *Gateway) registerConn(appID uuid.UUID, conn *websocket.Conn, handler *WebSocketHandler) {
	g.activeConnsMu.Lock()
	g.activeConns[appID] = &activeConn{conn: conn, handler: handler}
	g.activeConnsMu.Unlock()
}

// unregisterConn removes the tracked WebSocket connection for an application.
func (g *Gateway) unregisterConn(appID uuid.UUID) {
	g.activeConnsMu.Lock()
	delete(g.activeConns, appID)
	g.activeConnsMu.Unlock()
}

// getOrCreatePendingCounter returns the atomic counter for an app, creating it if needed.
func (g *Gateway) getOrCreatePendingCounter(appID uuid.UUID) *atomic.Int64 {
	key := appID.String()
	g.pendingCompletionsMu.Lock()
	defer g.pendingCompletionsMu.Unlock()
	if c, ok := g.pendingCompletions[key]; ok {
		return c
	}
	c := &atomic.Int64{}
	g.pendingCompletions[key] = c
	return c
}

// incPendingCompletion increments the in-flight file count for an app (call when queuing a job).
func (g *Gateway) incPendingCompletion(appID uuid.UUID) {
	g.getOrCreatePendingCounter(appID).Add(1)
}

// decPendingCompletion decrements the in-flight file count (call when a completion job finishes).
func (g *Gateway) decPendingCompletion(appID uuid.UUID) {
	g.getOrCreatePendingCounter(appID).Add(-1)
}

// waitForPendingCompletions blocks until the app has no pending file writes or timeout expires.
func (g *Gateway) waitForPendingCompletions(appID uuid.UUID, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	tick := pendingCompletionsTick()
	for time.Now().Before(deadline) {
		g.pendingCompletionsMu.RLock()
		c, ok := g.pendingCompletions[appID.String()]
		g.pendingCompletionsMu.RUnlock()
		if !ok || c.Load() <= 0 {
			return
		}
		time.Sleep(tick)
		if tick < 100*time.Millisecond {
			tick = 50 * time.Millisecond
		}
	}
	g.logger.Log(logger.Warning, "timeout waiting for pending file completions before build", appID.String())
}

// sendPipelineProgress sends a pipeline_progress message to the WebSocket client for the given app.
func (g *Gateway) sendPipelineProgress(appID uuid.UUID, stageId, message string) {
	g.activeConnsMu.RLock()
	ac := g.activeConns[appID]
	g.activeConnsMu.RUnlock()

	if ac == nil {
		return
	}

	msg := mover.SyncMessage{
		Type:      mover.MessageTypePipelineProgress,
		Timestamp: time.Now(),
		Payload: mover.PipelineProgressPayload{
			StageId: stageId,
			Message: message,
		},
	}
	if err := ac.handler.sendMessage(ac.conn, msg); err != nil {
		g.logger.Log(logger.Warning, "failed to send pipeline progress", fmt.Sprintf("app=%s err=%v", appID, err))
	}
}

// sendBuildStatus sends a build lifecycle status message to the WebSocket client for the given app.
func (g *Gateway) sendBuildStatus(appID uuid.UUID, phase, message, errMsg string) {
	g.activeConnsMu.RLock()
	ac := g.activeConns[appID]
	g.activeConnsMu.RUnlock()

	if ac == nil {
		return
	}

	msg := mover.SyncMessage{
		Type:      mover.MessageTypeBuildStatus,
		Timestamp: time.Now(),
		Payload: mover.BuildStatusPayload{
			Phase:   phase,
			Message: message,
			Error:   errMsg,
		},
	}
	if err := ac.handler.sendMessage(ac.conn, msg); err != nil {
		g.logger.Log(logger.Warning, "failed to send build status", fmt.Sprintf("app=%s err=%v", appID, err))
	}
}

// SendBuildLog streams a build log line to the WebSocket client for the given app.
func (g *Gateway) SendBuildLog(appID uuid.UUID, logLine string) {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	g.sendBuildLog(appID, logLine, timestamp)
}

// sendBuildLog sends a build log line to the WebSocket client for the given app.
func (g *Gateway) sendBuildLog(appID uuid.UUID, log string, timestamp string) {
	g.activeConnsMu.RLock()
	ac := g.activeConns[appID]
	g.activeConnsMu.RUnlock()

	if ac == nil {
		return
	}

	msg := mover.SyncMessage{
		Type:      mover.MessageTypeBuildLog,
		Timestamp: time.Now(),
		Payload: mover.BuildLogPayload{
			Log:       log,
			Timestamp: timestamp,
		},
	}
	if err := ac.handler.sendMessage(ac.conn, msg); err != nil {
		g.logger.Log(logger.Warning, "failed to send build log", fmt.Sprintf("app=%s err=%v", appID, err))
	}
}

// sendCodebaseIndexed sends codebase_indexed to the CLI when indexing is complete.
// Signals the CLI to run the deployment workflow.
func (g *Gateway) sendCodebaseIndexed(appCtx *ApplicationContext) error {
	g.activeConnsMu.RLock()
	ac := g.activeConns[appCtx.ApplicationID]
	g.activeConnsMu.RUnlock()

	if ac == nil {
		g.logger.Log(logger.Warning, "no active connection for codebase_indexed", appCtx.ApplicationID.String())
		return nil
	}

	payload := mover.CodebaseIndexedPayload{
		ApplicationID:  appCtx.ApplicationID.String(),
		OrganizationID: appCtx.OrganizationID.String(),
		Source:         appCtx.StagingPath,
		Mode:           "development",
	}
	msg := mover.SyncMessage{
		Type:      mover.MessageTypeCodebaseIndexed,
		Timestamp: time.Now(),
		Payload:   payload,
	}
	if err := ac.handler.sendMessage(ac.conn, msg); err != nil {
		g.logger.Log(logger.Warning, "failed to send codebase_indexed", fmt.Sprintf("app=%s err=%v", appCtx.ApplicationID, err))
		return err
	}
	g.logger.Log(logger.Info, "codebase_indexed sent", fmt.Sprintf("app=%s org=%s source=%s", appCtx.ApplicationID, appCtx.OrganizationID, appCtx.StagingPath))
	return nil
}

// sendDeploymentStatus sends a deployment status change to the WebSocket client for the given app.
// Called when a live_dev_status notification arrives from PostgreSQL.
func (g *Gateway) sendDeploymentStatus(appID uuid.UUID, status string) {
	g.activeConnsMu.RLock()
	ac := g.activeConns[appID]
	g.activeConnsMu.RUnlock()

	if ac == nil {
		return
	}

	msg := mover.SyncMessage{
		Type:      mover.MessageTypeDeploymentStatus,
		Timestamp: time.Now(),
		Payload: mover.DeploymentStatusPayload{
			Status: status,
		},
	}
	if err := ac.handler.sendMessage(ac.conn, msg); err != nil {
		g.logger.Log(logger.Warning, "failed to send deployment status", fmt.Sprintf("app=%s err=%v", appID, err))
	}
}

// HandleLiveDevNotification processes live_dev_logs and live_dev_status notifications
// from the PostgresListener. This is registered as a callback on the SocketServer.
func (g *Gateway) HandleLiveDevNotification(channel, payload string) {
	switch channel {
	case "live_dev_logs":
		var n struct {
			ApplicationID string `json:"application_id"`
			Log           string `json:"log"`
			CreatedAt     string `json:"created_at"`
		}
		if err := json.Unmarshal([]byte(payload), &n); err != nil {
			g.logger.Log(logger.Warning, "failed to parse live_dev_logs notification", err.Error())
			return
		}
		appID, err := uuid.Parse(n.ApplicationID)
		if err != nil {
			return
		}
		g.sendBuildLog(appID, n.Log, n.CreatedAt)

	case "live_dev_status":
		var n struct {
			ApplicationID string `json:"application_id"`
			Status        string `json:"status"`
		}
		if err := json.Unmarshal([]byte(payload), &n); err != nil {
			g.logger.Log(logger.Warning, "failed to parse live_dev_status notification", err.Error())
			return
		}
		appID, err := uuid.Parse(n.ApplicationID)
		if err != nil {
			return
		}
		g.sendDeploymentStatus(appID, n.Status)
	}
}

// completeFileSync performs write, manifest, chunk index, inject, and ACK.
// Used by workers and by sync fallback when completion queue is full.
// Staging write runs in parallel with DB ops; manifest and chunk index run sequentially
// to reduce peak DB connections (avoids exhausting Supabase connection pool).
// HandleFileWritten fires async.
// Returns error only when write fails (caller can send error to client); manifest/index failures are logged.
func (g *Gateway) completeFileSync(ctx context.Context, job *fileCompletionJob, conn *websocket.Conn, path string) error {
	var writeErr, manifestErr, indexErr error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		writeErr = WriteContentToStaging(ctx, job.stagingPath, job.path, job.content, job.checksum)
	}()
	go func() {
		defer wg.Done()
		manifestErr = AddPath(ctx, g.store, job.applicationID, job.path, job.checksum)
		indexErr = IndexFileChunks(ctx, g.store, job.applicationID, job.path, job.content)
	}()
	wg.Wait()

	if writeErr != nil {
		g.logger.Log(logger.Error, "failed to write file to staging", fmt.Sprintf("path=%s err=%v", job.path, writeErr))
		return fmt.Errorf("failed to write file to staging: %w", writeErr)
	}
	if manifestErr != nil {
		g.logger.Log(logger.Warning, "failed to persist manifest", fmt.Sprintf("app=%s err=%v", job.applicationID, manifestErr))
	}
	if indexErr != nil {
		g.logger.Log(logger.Warning, "failed to index file chunks", fmt.Sprintf("app=%s path=%s err=%v", job.applicationID, job.path, indexErr))
	}

	g.logger.Log(logger.Debug, "file received and written", job.path)

	if g.buildFirstManager != nil {
		go g.buildFirstManager.HandleFileWritten(ctx, job.appCtx, job.path, job.content)
	}

	if err := g.websocketHandler.sendAck(conn, path); err != nil {
		g.logger.Log(logger.Warning, "failed to send ACK", fmt.Sprintf("path=%s err=%v", path, err))
	}
	return nil
}

// runFileCompletionWorker processes file write + ACK. Delegates to completeFileSync for
// parallel write, manifest, chunk index, and async HandleFileWritten.
func (g *Gateway) runFileCompletionWorker() {
	for job := range g.completionJobs {
		ctx := context.WithValue(context.Background(), shared_types.OrganizationIDKey, job.appCtx.OrganizationID.String())
		g.completeFileSync(ctx, job, job.conn, job.path)
		g.decPendingCompletion(job.applicationID)
		putJob(job)
	}
}

// BuildFirstManager returns the build-first manager for wiring callbacks (e.g. from TaskService).
func (g *Gateway) BuildFirstManager() *BuildFirstManager {
	return g.buildFirstManager
}

// GetSessionEnv returns env vars sent from client (set-env file values) for an application.
// Returns a copy to prevent callers from mutating the shared store.
func (g *Gateway) GetSessionEnv(applicationID uuid.UUID) map[string]string {
	g.sessionEnvStoreMu.RLock()
	env := g.sessionEnvStore[applicationID.String()]
	if env == nil {
		g.sessionEnvStoreMu.RUnlock()
		return nil
	}
	// Copy while holding RLock so we get a consistent snapshot
	out := make(map[string]string, len(env))
	for k, v := range env {
		out[k] = v
	}
	g.sessionEnvStoreMu.RUnlock()
	return out
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

// handleMessage routes messages to appropriate handlers based on message type.
// msg.Payload is json.RawMessage — handlers unmarshal directly (no double encode).
func (g *Gateway) handleMessage(ctx context.Context, conn *websocket.Conn, appCtx *ApplicationContext, msg *recvMessage) error {
	switch msg.Type {
	case mover.MessageTypeFileChange:
		return g.handleFileChange(appCtx, msg.Payload)
	case mover.MessageTypeFileContent:
		return g.handleFileContent(ctx, conn, appCtx, msg.Payload)
	case mover.MessageTypeFileDelete:
		return g.handleFileDelete(ctx, appCtx, msg.Payload)
	case mover.MessageTypeEnvVars:
		return g.handleEnvVars(ctx, appCtx, msg.Payload)
	case mover.MessageTypeSyncComplete:
		return g.handleSyncComplete(ctx, appCtx)
	case mover.MessageTypeTriggerBuild:
		return g.handleTriggerBuild(ctx, appCtx, msg.Payload)
	case mover.MessageTypePing:
		return g.websocketHandler.sendPong(conn)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
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

func (g *Gateway) handleSyncComplete(ctx context.Context, appCtx *ApplicationContext) error {
	g.logger.Log(logger.Info, "sync_complete received", appCtx.ApplicationID.String())
	g.waitForPendingCompletions(appCtx.ApplicationID, 60*time.Second)

	if g.buildFirstManager.TryRecoverFromSyncComplete(ctx, appCtx) {
		return nil
	}

	return g.sendCodebaseIndexed(appCtx)
}

func (g *Gateway) handleTriggerBuild(ctx context.Context, appCtx *ApplicationContext, payload json.RawMessage) error {
	var p mover.TriggerBuildPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("failed to unmarshal trigger_build payload: %w", err)
	}
	if p.Dockerfile == "" {
		return fmt.Errorf("trigger_build requires dockerfile")
	}

	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, appCtx.OrganizationID.String())
	if err := g.buildFirstManager.StartBuildFromDockerfile(orgCtx, appCtx, p.Dockerfile, p.Dockerignore, p.Port, p.Workdir); err != nil {
		return err
	}
	return nil
}

func (g *Gateway) handleEnvVars(ctx context.Context, appCtx *ApplicationContext, payload json.RawMessage) error {
	var p mover.EnvVarsPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("failed to unmarshal env_vars payload: %w", err)
	}
	if len(p.Vars) == 0 {
		return nil
	}
	// Store a copy to prevent client from mutating our stored map
	vars := make(map[string]string, len(p.Vars))
	for k, v := range p.Vars {
		vars[k] = v
	}
	g.sessionEnvStoreMu.Lock()
	g.sessionEnvStore[appCtx.ApplicationID.String()] = vars
	g.sessionEnvStoreMu.Unlock()
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, appCtx.OrganizationID.String())
	if err := tasks.UpdateLiveDevServiceEnv(orgCtx, appCtx.ApplicationID, p.Vars); err != nil {
		g.logger.Log(logger.Warning, "failed to update service env", fmt.Sprintf("app=%s err=%v", appCtx.ApplicationID, err))
	} else {
		g.logger.Log(logger.Info, "env vars updated, service rolling new tasks", appCtx.ApplicationID.String())
	}
	return nil
}

func (g *Gateway) handleFileChange(appCtx *ApplicationContext, payload json.RawMessage) error {
	var fileChange mover.FileChange
	if err := json.Unmarshal(payload, &fileChange); err != nil {
		return fmt.Errorf("failed to unmarshal file_change payload: %w", err)
	}

	// Validate file path is within base_path
	// Paths from client are relative to base_path, so we validate they don't escape
	if err := g.validateFilePath(fileChange.Path, appCtx.BasePath); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	totalChunks := int((fileChange.Size + chunkSize() - 1) / chunkSize())
	if totalChunks == 0 {
		totalChunks = 1
	}
	g.stagingManager.GetFileReceiver(appCtx.ApplicationID, fileChange.Path, totalChunks, fileChange.Checksum, appCtx.StagingPath)

	g.logger.Log(logger.Debug, "file change received", fmt.Sprintf("path=%s op=%s size=%d chunks=%d", fileChange.Path, fileChange.Operation, fileChange.Size, totalChunks))
	return nil
}

func (g *Gateway) handleFileContent(ctx context.Context, conn *websocket.Conn, appCtx *ApplicationContext, payload json.RawMessage) error {
	var fileContent mover.FileContent
	if err := json.Unmarshal(payload, &fileContent); err != nil {
		return fmt.Errorf("failed to unmarshal file_content payload: %w", err)
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
	job := getJob()
	job.content = content
	job.path = fileContent.Path
	job.checksum = fileContent.Checksum
	job.stagingPath = appCtx.StagingPath
	job.appCtx = appCtx
	job.conn = conn
	job.applicationID = appCtx.ApplicationID

	// During shutdown, run inline to avoid send on closed channel
	if g.shuttingDown.Load() {
		err := g.completeFileSync(ctx, job, conn, fileContent.Path)
		putJob(job)
		return err
	}

	select {
	case g.completionJobs <- job:
		g.incPendingCompletion(appCtx.ApplicationID)
		// Queued - worker will handle it and putJob
	default:
		// Queue full - fall back to parallel sync to avoid dropping
		if err := g.completeFileSync(ctx, job, conn, fileContent.Path); err != nil {
			putJob(job)
			return err
		}
		putJob(job)
	}

	return nil
}

func (g *Gateway) handleFileDelete(ctx context.Context, appCtx *ApplicationContext, payload json.RawMessage) error {
	var fileChange mover.FileChange
	if err := json.Unmarshal(payload, &fileChange); err != nil {
		return fmt.Errorf("failed to unmarshal file_delete payload: %w", err)
	}

	// Validate file path is within base_path
	// Paths from client are relative to base_path, so we validate they don't escape
	if err := g.validateFilePath(fileChange.Path, appCtx.BasePath); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	if err := DeleteFileFromStaging(ctx, appCtx.StagingPath, fileChange.Path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if err := RemovePath(ctx, g.store, appCtx.ApplicationID, fileChange.Path); err != nil {
		g.logger.Log(logger.Warning, "failed to persist manifest after delete", fmt.Sprintf("app=%s err=%v", appCtx.ApplicationID, err))
	}
	if err := DeleteFileChunks(ctx, g.store, appCtx.ApplicationID, fileChange.Path); err != nil {
		g.logger.Log(logger.Warning, "failed to delete file chunks", fmt.Sprintf("app=%s path=%s err=%v", appCtx.ApplicationID, fileChange.Path, err))
	}

	g.logger.Log(logger.Debug, "file deleted", fileChange.Path)
	if g.buildFirstManager != nil {
		g.buildFirstManager.HandleFileDeleted(ctx, appCtx, fileChange.Path)
	}
	return nil
}
