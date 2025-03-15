package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// DeleteDomain deletes an existing domain in the application.
//
// It takes a domain ID and deletes the associated domain in the storage layer.
// If the domain does not exist, it returns ErrDomainNotFound.
// If any error occurs while deleting the domain, it returns the error.
func (s *DomainsService) DeleteDomain(domainID string) error {
	s.logger.Log(logger.Debug, "delete domain request received: domain_id=%s\n", domainID)

	_, err := uuid.Parse(domainID)
	if err != nil {
		return types.ErrInvalidDomainID
	}

	existing_domain, err := s.storage.GetDomain(domainID)
	if err != nil {
		return err
	}

	if existing_domain == nil {
		return types.ErrDomainNotFound
	}

	if err := s.storage.DeleteDomain(existing_domain); err != nil {
		return err
	}

	return nil
}
