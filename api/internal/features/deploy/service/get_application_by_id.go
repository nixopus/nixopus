package service

import (
	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *DeployService) GetApplicationById(id string, organizationID uuid.UUID) (shared_types.Application, error) {
	return s.storage.GetApplicationById(id, organizationID)
}
