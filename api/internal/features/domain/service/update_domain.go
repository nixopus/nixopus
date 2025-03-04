package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
)

// UpdateDomain updates an existing domain in the application.
//
// It takes a domain ID, a userID and a new name and updates the domain in the storage layer.
// If the domain does not exist, it returns ErrDomainNotFound.
// If any error occurs while updating the domain, it returns the error.
func (s *DomainsService) UpdateDomain(newName string, userID string, domainID string) (types.CreateDomainResponse, error) {
	existing_domain, err := s.storage.GetDomain(domainID)
	if err != nil {
		return types.CreateDomainResponse{}, err
	}
	if existing_domain == nil {
		return types.CreateDomainResponse{}, types.ErrDomainNotFound
	}

	if err := s.storage.UpdateDomain(existing_domain.ID.String(), newName); err != nil {
		return types.CreateDomainResponse{}, err
	}

	return types.CreateDomainResponse{}, nil
}
