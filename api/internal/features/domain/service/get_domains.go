package service

import (
	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetDomains retrieves all domains from the storage.
//
// This method calls the storage layer to fetch the complete list of domains.
// It returns the list of domains or an error if fetching fails.
//
// Returns:
//
//	([]shared_types.Domain, error) - A slice of Domain objects and an error if any occurred.
func (s *DomainsService) GetDomains(organization_id string, UserID uuid.UUID) ([]shared_types.Domain, error) {
	return s.storage.GetDomains(organization_id, UserID)
}
