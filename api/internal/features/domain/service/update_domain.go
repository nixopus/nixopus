package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// UpdateDomain updates an existing domain in the application.
//
// It takes a domain ID, a userID and a new name and updates the domain in the storage layer.
// If the domain does not exist, it returns ErrDomainNotFound.
// If any error occurs while updating the domain, it returns the error.
// UpdateDomain updates a domain's name.
//
// It takes a new name, user ID, and domain ID, then updates the domain in the storage layer.
// Returns the updated domain and any error that occurred.
func (s *DomainsService) UpdateDomain(newName, userID, domainID string) (*shared_types.Domain, error) {
	s.logger.Log(logger.Debug, fmt.Sprintf("update domain request received: domain_id=%s, user_id=%s, new_name=%s\n", domainID, userID, newName), "")

	_, err := uuid.Parse(domainID)
	if err != nil {
		return nil, types.ErrInvalidDomainID
	}

	existing, err := s.storage.GetDomain(domainID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, types.ErrDomainNotFound
	}

	err = s.storage.UpdateDomain(domainID, newName)
	if err != nil {
		return nil, err
	}

	updated, err := s.storage.GetDomain(domainID)
	if err != nil {
		return nil, err
	}

	return updated, nil
}
