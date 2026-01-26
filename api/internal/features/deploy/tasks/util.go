package tasks

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func GetStringFromMap(m map[string]string) string {
	var result string
	for key, value := range m {
		result += key + "=" + value + " "
	}
	return result
}

func GetMapFromString(s string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(s, " ")
	for _, pair := range pairs {
		if pair != "" {
			kv := strings.Split(pair, "=")
			if len(kv) == 2 {
				result[kv[0]] = kv[1]
			}
		}
	}
	return result
}

type TaskContext struct {
	service       *TaskService
	applicationID uuid.UUID
	deploymentID  uuid.UUID
	statusID      uuid.UUID
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
		tc.service.Logger.Log(logger.Error, "Failed to add application log: "+err.Error(), "")
	}
}

func (tc *TaskContext) LogAndUpdateStatus(message string, status shared_types.Status) {
	tc.AddLog(message)
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

	return &LiveDevTaskContext{
		service:       s,
		applicationID: config.ApplicationID,
		deploymentID:  deploymentID,
		statusID:      statusID,
	}, nil
}

// AddLog adds a log entry for the live dev deployment
func (tc *LiveDevTaskContext) AddLog(logMessage string) {
	appLog := shared_types.ApplicationLogs{
		ID:                      uuid.New(),
		ApplicationID:           tc.applicationID,
		Log:                     "[LiveDev] " + logMessage,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		ApplicationDeploymentID: tc.deploymentID,
	}

	err := tc.service.Storage.AddApplicationLogs(&appLog)
	if err != nil {
		tc.service.Logger.Log(logger.Error, "Failed to add live dev log: "+err.Error(), "")
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

// GetDeploymentID returns the deployment ID
func (tc *LiveDevTaskContext) GetDeploymentID() uuid.UUID {
	return tc.deploymentID
}
