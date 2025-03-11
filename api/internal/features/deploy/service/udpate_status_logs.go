package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// updateStatus updates the application status
func (s *DeployService) updateStatus(deploymentID uuid.UUID, status shared_types.Status, id uuid.UUID) {
	appStatus := shared_types.ApplicationDeploymentStatus{
		ID:                      id,
		ApplicationDeploymentID: deploymentID,
		Status:                  status,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	err := s.storage.UpdateApplicationDeploymentStatus(&appStatus)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to update application deployment status: "+err.Error(), "")
	}
}

// addLog adds a new log entry for the application
func (s *DeployService) addLog(applicationID uuid.UUID, logMessage string, deploymentID uuid.UUID) {
	appLog := shared_types.ApplicationLogs{
		ID:                      uuid.New(),
		ApplicationID:           applicationID,
		Log:                     logMessage,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		ApplicationDeploymentID: deploymentID,
	}

	err := s.storage.AddApplicationLogs(&appLog)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to add application log: "+err.Error(), "")
	}
}

func GetStringFromMap(m map[string]string) string {
	var result string
	for key, value := range m {
		result += key + "=" + value + " "
	}
	return result
}
