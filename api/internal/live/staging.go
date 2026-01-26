package live

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	github_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// StagingManager manages staging directories for live deployments using the same pattern as normal deployments
type StagingManager struct {
	githubService *github_service.GithubConnectorService
	deployStorage storage.DeployRepository
	fileReceivers map[uuid.UUID]map[string]*FileReceiver // applicationID -> path -> FileReceiver
	mu            sync.RWMutex
}

// NewStagingManager creates a new staging manager
func NewStagingManager(githubService *github_service.GithubConnectorService, deployStorage storage.DeployRepository) *StagingManager {
	return &StagingManager{
		githubService: githubService,
		deployStorage: deployStorage,
		fileReceivers: make(map[uuid.UUID]map[string]*FileReceiver),
	}
}

// GetStagingPath gets or creates the staging path for an application using the same pattern as normal deployments
// This reuses GetClonePath from github-connector service
// For monorepo apps, the staging path includes the base_path subdirectory
func (sm *StagingManager) GetStagingPath(ctx context.Context, applicationID, userID, organizationID uuid.UUID) (string, error) {
	// Get application to get environment and base_path
	application, err := sm.deployStorage.GetApplicationById(applicationID.String(), organizationID)
	if err != nil {
		return "", err
	}

	// Use GetClonePath which creates the staging directory using the same pattern as normal deployments
	// Path structure: {mountPath}/{userID}/{environment}/{applicationID}
	stagingPath, _, err := sm.githubService.GetClonePath(userID.String(), string(application.Environment), applicationID.String())
	if err != nil {
		return "", err
	}

	// Add base_path to staging path if set (for monorepo apps)
	// If base_path is "/" or empty, use the root staging path
	if application.BasePath != "" && application.BasePath != "/" {
		stagingPath = filepath.Join(stagingPath, application.BasePath)
	}

	return stagingPath, nil
}

// GetFileReceiver gets or creates a file receiver for an application and path
func (sm *StagingManager) GetFileReceiver(applicationID uuid.UUID, path string, totalChunks int, checksum string, stagingPath string) *FileReceiver {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.fileReceivers[applicationID] == nil {
		sm.fileReceivers[applicationID] = make(map[string]*FileReceiver)
	}

	if receiver, exists := sm.fileReceivers[applicationID][path]; exists {
		if receiver.Checksum != checksum {
			receiver.Reset(totalChunks, checksum)
		}
		return receiver
	}

	receiver := NewFileReceiver(path, totalChunks, checksum, stagingPath)
	sm.fileReceivers[applicationID][path] = receiver
	return receiver
}

// RemoveFileReceiver removes a file receiver
func (sm *StagingManager) RemoveFileReceiver(applicationID uuid.UUID, path string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if receivers, exists := sm.fileReceivers[applicationID]; exists {
		delete(receivers, path)
		if len(receivers) == 0 {
			delete(sm.fileReceivers, applicationID)
		}
	}
}

// GetStagingPathForApplication is a helper that gets staging path from application
func GetStagingPathForApplication(mountPath string, userID uuid.UUID, environment shared_types.Environment, applicationID uuid.UUID) string {
	return filepath.Join(mountPath, userID.String(), string(environment), applicationID.String())
}
