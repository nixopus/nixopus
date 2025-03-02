package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
)

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
