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
