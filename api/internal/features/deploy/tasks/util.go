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
