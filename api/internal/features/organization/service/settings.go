package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// GetOrganizationSettings retrieves organization settings
func (s *OrganizationService) GetOrganizationSettings(organizationID string) (*types.OrganizationSettings, error) {
	s.logger.Log(logger.Info, "getting organization settings", "")

	// Type assert to get access to settings methods
	store, ok := s.storage.(*storage.OrganizationStore)
	if !ok {
		return nil, types.ErrFailedToGetOrganizationFromContext
	}
	return store.GetOrganizationSettings(organizationID)
}

// UpdateOrganizationSettings updates organization settings with the provided data
func (s *OrganizationService) UpdateOrganizationSettings(organizationID string, settings types.OrganizationSettingsData) (*types.OrganizationSettings, error) {
	s.logger.Log(logger.Info, "updating organization settings", "")

	// Type assert to get access to settings methods
	store, ok := s.storage.(*storage.OrganizationStore)
	if !ok {
		return nil, types.ErrFailedToGetOrganizationFromContext
	}
	return store.UpdateOrganizationSettings(organizationID, settings)
}
