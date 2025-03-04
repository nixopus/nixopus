package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
)

// DeleteDomain deletes an existing domain in the application.
//
// It takes a domain ID and deletes the associated domain in the storage layer.
// If the domain does not exist, it returns ErrDomainAlreadyExists.
// If any error occurs while deleting the domain, it returns the error.
func (s *DomainsService) DeleteDomain(domainID string) error {
	existing_domain, err := s.storage.GetDomain(domainID)
	if existing_domain.ID == uuid.Nil {
		return types.ErrDomainAlreadyExists
	}

	if err != nil {
		return err
	}

	if err := s.storage.DeleteDomain(existing_domain); err != nil {
		return err
	}

	return nil
}
