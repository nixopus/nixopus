package service

import (
	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *DeployService) GetDeploymentById(deploymentID string) (shared_types.ApplicationDeployment, error) {
	return s.storage.GetApplicationDeploymentById(deploymentID)
}

func (s *DeployService) GetApplicationDeployments(applicationID uuid.UUID, page, pageSize int) ([]shared_types.ApplicationDeployment, int, error) {
	return s.storage.GetPaginatedApplicationDeployments(applicationID, page, pageSize)
}

func (s *DeployService) GetLatestDeployments(organizationID string, limit int) ([]shared_types.ApplicationDeployment, error) {
	orgUUID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, err
	}
	return s.storage.GetLatestDeployments(orgUUID, limit)
}
