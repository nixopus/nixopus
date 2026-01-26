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

	"github.com/gorilla/websocket"
	api_key_service "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	api_key_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/mover"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

const chunkSize = int64(64 * 1024)

type Gateway struct {
	stagingManager   *StagingManager
	serviceManager   *ServiceManager
	websocketHandler *WebSocketHandler
	store            *shared_storage.Store
	logger           logger.Logger
}

func NewGateway(stagingManager *StagingManager, taskService *tasks.TaskService, store *shared_storage.Store) *Gateway {
	logger := logger.NewLogger()
	gateway := &Gateway{
		stagingManager: stagingManager,
		serviceManager: NewServiceManager(stagingManager, taskService, logger),
		store:          store,
		logger:         logger,
	}
	gateway.websocketHandler = NewWebSocketHandler(gateway, logger)
	return gateway
}

// HandleWebSocket delegates to the WebSocket handler
func (g *Gateway) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	g.websocketHandler.HandleWebSocket(w, r)
}

func (g *Gateway) verifyAPIKey(ctx context.Context, token string) (*shared_types.APIKey, error) {
	apiKeyStorage := api_key_storage.APIKeyStorage{DB: g.store.DB, Ctx: ctx}
	apiKeyService := api_key_service.NewAPIKeyService(apiKeyStorage, g.logger)
	return apiKeyService.VerifyAPIKey(token)
}

// handleMessage routes messages to appropriate handlers based on message type
func (g *Gateway) handleMessage(ctx context.Context, conn *websocket.Conn, appCtx *ApplicationContext, msg *mover.SyncMessage) error {
	switch msg.Type {
	case mover.MessageTypeFileChange:
		return g.handleFileChange(appCtx, msg)
	case mover.MessageTypeFileContent:
		return g.handleFileContent(ctx, conn, appCtx, msg)
	case mover.MessageTypeFileDelete:
		return g.handleFileDelete(appCtx, msg)
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

	// Write to staging
	if err := receiver.WriteToStaging(); err != nil {
		return fmt.Errorf("failed to write file to staging: %w", err)
	}

	g.logger.Log(logger.Info, "file received and written", fileContent.Path)
	g.stagingManager.RemoveFileReceiver(appCtx.ApplicationID, fileContent.Path)

	if g.serviceManager != nil {
		g.serviceManager.EnsureDevServiceStarted(ctx, appCtx)
	}

	if err := g.websocketHandler.sendAck(conn, fileContent.Path); err != nil {
		g.logger.Log(logger.Warning, "failed to send ACK", fmt.Sprintf("path=%s err=%v", fileContent.Path, err))
	}

	return nil
}

func (g *Gateway) handleFileDelete(appCtx *ApplicationContext, msg *mover.SyncMessage) error {
	var fileChange mover.FileChange
	if err := g.unmarshalPayload(msg.Payload, &fileChange); err != nil {
		return err
	}

	// Validate file path is within base_path
	// Paths from client are relative to base_path, so we validate they don't escape
	if err := g.validateFilePath(fileChange.Path, appCtx.BasePath); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	if err := DeleteFileFromStaging(appCtx.StagingPath, fileChange.Path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	g.logger.Log(logger.Info, "file deleted", fileChange.Path)
	return nil
}
