package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
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
	ctx := fmt.Sprintf("user_id=%s org_id=%s", UserID, organization_id)
	s.logger.Log(logger.Debug, "get domains: storage lookup", ctx)

	domains, err := s.storage.GetDomains(organization_id, UserID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("get domains: %v", err), ctx)
		return nil, err
	}

	s.logger.Log(logger.Info, "get domains: storage success", fmt.Sprintf("%s count=%d", ctx, len(domains)))
	return domains, nil
}
