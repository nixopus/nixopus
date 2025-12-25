package service

import (
	"github.com/google/uuid"
)

func (s *DeployService) UpdateApplicationLabels(applicationID uuid.UUID, labels []string, organizationID uuid.UUID) error {
	return s.storage.UpdateApplicationLabels(applicationID, labels, organizationID)
}
