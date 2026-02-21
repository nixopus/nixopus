package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetStringFromMap serializes a map to a JSON string for safe storage.
// Falls back to legacy space-delimited format only if JSON marshaling fails.
func GetStringFromMap(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	data, err := json.Marshal(m)
	if err != nil {
		var result string
		for key, value := range m {
			result += key + "=" + value + " "
		}
		return result
	}
	return string(data)
}

// GetMapFromString deserializes a string to a map. Supports both JSON format
// and the legacy space-delimited key=value format for backward compatibility.
func GetMapFromString(s string) map[string]string {
	if s == "" {
		return make(map[string]string)
	}
	result := make(map[string]string)
	if err := json.Unmarshal([]byte(s), &result); err == nil {
		return result
	}
	// Legacy format: space-delimited key=value pairs
	pairs := strings.Split(s, " ")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		idx := strings.IndexByte(pair, '=')
		if idx > 0 {
			result[pair[:idx]] = pair[idx+1:]
		}
	}
	return result
}

const logBatchSize = 50

type TaskContext struct {
	mu            sync.Mutex
	service       *TaskService
	applicationID uuid.UUID
	deploymentID  uuid.UUID
	statusID      uuid.UUID
	onLogCallback func(applicationID uuid.UUID, logLine string) // for live dev real-time streaming
	logBuffer     []shared_types.ApplicationLogs
}

func (s *TaskService) NewTaskContext(result shared_types.TaskPayload) *TaskContext {
	var statusID uuid.UUID
	if result.Status != nil {
		statusID = result.Status.ID
	} else {
		statusID = uuid.New()
	}

	return &TaskContext{
		service:       s,
		applicationID: result.Application.ID,
		deploymentID:  result.ApplicationDeployment.ID,
		statusID:      statusID,
		logBuffer:     make([]shared_types.ApplicationLogs, 0, logBatchSize),
	}
}

func (tc *TaskContext) UpdateDeployment(deployment *shared_types.ApplicationDeployment) {
	err := tc.service.Storage.UpdateApplicationDeployment(deployment)
	if err != nil {
		tc.service.Logger.Log(logger.Error, "Failed to update application deployment: "+err.Error(), "")
	}
}

func (tc *TaskContext) UpdateStatus(status shared_types.Status) {
	appStatus := shared_types.ApplicationDeploymentStatus{
		ID:                      tc.statusID,
		ApplicationDeploymentID: tc.deploymentID,
		Status:                  status,
		UpdatedAt:               time.Now(),
	}

	err := tc.service.Storage.UpdateApplicationDeploymentStatus(&appStatus)
	if err != nil {
		tc.service.Logger.Log(logger.Error, "Failed to update application deployment status: "+err.Error(), "")
	}
}

func (tc *TaskContext) AddLog(logMessage string) {
	tc.mu.Lock()
	appLog := shared_types.ApplicationLogs{
		ID:                      uuid.New(),
		ApplicationID:           tc.applicationID,
		Log:                     logMessage,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		ApplicationDeploymentID: tc.deploymentID,
	}

	tc.logBuffer = append(tc.logBuffer, appLog)
	needsFlush := len(tc.logBuffer) >= logBatchSize
	tc.mu.Unlock()

	if needsFlush {
		tc.FlushLogs()
	}
	if tc.onLogCallback != nil {
		tc.onLogCallback(tc.applicationID, logMessage)
	}
}

// FlushLogs writes any buffered logs to the database in a single batch INSERT.
func (tc *TaskContext) FlushLogs() {
	tc.mu.Lock()
	if len(tc.logBuffer) == 0 {
		tc.mu.Unlock()
		return
	}
	batch := make([]shared_types.ApplicationLogs, len(tc.logBuffer))
	copy(batch, tc.logBuffer)
	tc.logBuffer = tc.logBuffer[:0]
	tc.mu.Unlock()

	if err := tc.service.Storage.AddApplicationLogsBatch(batch); err != nil {
		tc.service.Logger.Log(logger.Error, "Failed to flush log batch: "+err.Error(), "")
		for _, log := range batch {
			if singleErr := tc.service.Storage.AddApplicationLogs(&log); singleErr != nil {
				tc.service.Logger.Log(logger.Error, "Failed to add individual log: "+singleErr.Error(), "")
			}
		}
	}
}

func (tc *TaskContext) LogAndUpdateStatus(message string, status shared_types.Status) {
	tc.AddLog(message)
	tc.FlushLogs()
	tc.UpdateStatus(status)
}

func (tc *TaskContext) GetApplicationID() uuid.UUID {
	return tc.applicationID
}

func (tc *TaskContext) GetDeploymentID() uuid.UUID {
	return tc.deploymentID
}

func (tc *TaskContext) GetStatusID() uuid.UUID {
	return tc.statusID
}

// LiveDevTaskContext provides logging context for live dev deployments
type LiveDevTaskContext struct {
	service       *TaskService
	applicationID uuid.UUID
	deploymentID  uuid.UUID
	statusID      uuid.UUID
	// OnBuildLog is called for every log line during the build, enabling real-time
	// streaming to the CLI via WebSocket. When nil, logs are only written to the DB.
	OnBuildLog func(applicationID uuid.UUID, logLine string)
}

// NewLiveDevTaskContext creates a new task context for live dev deployments
// It creates a deployment record to enable logging using the existing infrastructure
func (s *TaskService) NewLiveDevTaskContext(config LiveDevConfig) (*LiveDevTaskContext, error) {
	// Create a deployment record for this live dev session
	deploymentID := uuid.New()
	statusID := uuid.New()

	deployment := &shared_types.ApplicationDeployment{
		ID:              deploymentID,
		ApplicationID:   config.ApplicationID,
		CommitHash:      "live-dev-" + config.ApplicationID.String()[:8],
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ContainerID:     "",
		ContainerName:   "nixopus-dev-" + config.ApplicationID.String(),
		ContainerImage:  "",
		ContainerStatus: "starting",
	}

	if err := s.Storage.AddApplicationDeployment(deployment); err != nil {
		s.Logger.Log(logger.Error, "Failed to create live dev deployment record: "+err.Error(), "")
		return nil, err
	}

	// Create initial status
	initialStatus := &shared_types.ApplicationDeploymentStatus{
		ID:                      statusID,
		ApplicationDeploymentID: deploymentID,
		Status:                  shared_types.Started,
		UpdatedAt:               time.Now(),
	}

	if err := s.Storage.AddApplicationDeploymentStatus(initialStatus); err != nil {
		s.Logger.Log(logger.Error, "Failed to create live dev deployment status: "+err.Error(), "")
		return nil, err
	}

	tc := &LiveDevTaskContext{
		service:       s,
		applicationID: config.ApplicationID,
		deploymentID:  deploymentID,
		statusID:      statusID,
	}
	if s.OnLiveDevLog != nil {
		tc.OnBuildLog = s.OnLiveDevLog
	}
	return tc, nil
}

// AddLog adds a log entry for the live dev deployment and streams to CLI if OnBuildLog is set.
func (tc *LiveDevTaskContext) AddLog(logMessage string) {
	appLog := shared_types.ApplicationLogs{
		ID:                      uuid.New(),
		ApplicationID:           tc.applicationID,
		Log:                     logMessage,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		ApplicationDeploymentID: tc.deploymentID,
	}

	err := tc.service.Storage.AddApplicationLogs(&appLog)
	if err != nil {
		tc.service.Logger.Log(logger.Error, "Failed to add live dev log: "+err.Error(), "")
	}
	if tc.OnBuildLog != nil {
		tc.OnBuildLog(tc.applicationID, logMessage)
	}
}

// UpdateStatus updates the deployment status for the live dev session
func (tc *LiveDevTaskContext) UpdateStatus(status shared_types.Status) {
	appStatus := shared_types.ApplicationDeploymentStatus{
		ID:                      tc.statusID,
		ApplicationDeploymentID: tc.deploymentID,
		Status:                  status,
		UpdatedAt:               time.Now(),
	}

	err := tc.service.Storage.UpdateApplicationDeploymentStatus(&appStatus)
	if err != nil {
		tc.service.Logger.Log(logger.Error, "Failed to update live dev deployment status: "+err.Error(), "")
	}
}

// LogAndUpdateStatus logs a message and updates the deployment status
func (tc *LiveDevTaskContext) LogAndUpdateStatus(message string, status shared_types.Status) {
	tc.AddLog(message)
	tc.UpdateStatus(status)
}

// UpdateDeployment updates the deployment record
func (tc *LiveDevTaskContext) UpdateDeployment(updates map[string]interface{}) {
	deployment := &shared_types.ApplicationDeployment{
		ID:        tc.deploymentID,
		UpdatedAt: time.Now(),
	}

	if containerID, ok := updates["container_id"].(string); ok {
		deployment.ContainerID = containerID
	}
	if containerName, ok := updates["container_name"].(string); ok {
		deployment.ContainerName = containerName
	}
	if containerImage, ok := updates["container_image"].(string); ok {
		deployment.ContainerImage = containerImage
	}
	if containerStatus, ok := updates["container_status"].(string); ok {
		deployment.ContainerStatus = containerStatus
	}

	err := tc.service.Storage.UpdateApplicationDeployment(deployment)
	if err != nil {
		tc.service.Logger.Log(logger.Error, "Failed to update live dev deployment: "+err.Error(), "")
	}
}

func (tc *LiveDevTaskContext) GetDeploymentID() uuid.UUID {
	return tc.deploymentID
}

func (tc *LiveDevTaskContext) toTaskContext() *TaskContext {
	var onLog func(applicationID uuid.UUID, logLine string)
	if tc.service.OnLiveDevLog != nil {
		onLog = tc.service.OnLiveDevLog
	}
	return &TaskContext{
		service:       tc.service,
		applicationID: tc.applicationID,
		deploymentID:  tc.deploymentID,
		statusID:      tc.statusID,
		onLogCallback: onLog,
	}
}

func GetSSHHostForOrganization(ctx context.Context, organizationID uuid.UUID) (string, error) {
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, organizationID.String())
	manager, err := ssh.GetSSHManagerFromContext(orgCtx)
	if err != nil {
		return "", fmt.Errorf("failed to get SSH manager: %w", err)
	}
	upstreamHost, err := manager.GetSSHHost()
	if err != nil {
		return "", fmt.Errorf("failed to get SSH host: %w", err)
	}
	return upstreamHost, nil
}
