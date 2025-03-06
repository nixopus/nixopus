package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"time"
)

func (s *DeployService) CreateDeployment(deployment *types.CreateDeploymentRequest, userID uuid.UUID) error {
	application := shared_types.Application{
		ID:                   uuid.New(),
		Name:                 deployment.Name,
		BuildVariables:       GetStringFromMap(deployment.BuildVariables),
		EnvironmentVariables: GetStringFromMap(deployment.EnvironmentVariables),
		Environment:          deployment.Environment,
		BuildPack:            deployment.BuildPack,
		Repository:           deployment.Repository,
		Branch:               deployment.Branch,
		PreRunCommand:        deployment.PreRunCommand,
		PostRunCommand:       deployment.PostRunCommand,
		Port:                 deployment.Port,
		DomainID:             deployment.DomainID,
		UserID:               userID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	application_status := shared_types.ApplicationStatus{
		ID:            uuid.New(),
		ApplicationID: application.ID,
		Status:        shared_types.Started,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	application_logs := shared_types.ApplicationLogs{
		ID:            uuid.New(),
		ApplicationID: application.ID,
		Log:           "Deployment started",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.storage.AddApplication(&application)

	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}

	err = s.storage.AddApplicationStatus(&application_status)

	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}

	err = s.storage.AddApplicationLogs(&application_logs)

	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}

	return nil
}

func GetStringFromMap(m map[string]string) string {
	var result string
	for key, value := range m {
		result += key + "=" + value + " "
	}
	return result
}
