package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
)

func (s *DomainsService) UpdateDomain(newName string,userID string,domainID string) (types.CreateDomainResponse, error) {
	existing_domain, err := s.storage.GetDomain(domainID)
	if err != nil {
		return types.CreateDomainResponse{}, err
	}
	if existing_domain == nil {
		return types.CreateDomainResponse{}, types.ErrDomainNotFound
	}

	if err := s.storage.UpdateDomain(existing_domain.ID.String(),newName); err != nil {
		return types.CreateDomainResponse{}, err
	}

	return types.CreateDomainResponse{}, nil
}
