package service

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *DomainsService) GetDomains() ([]shared_types.Domain, error) {
	return s.storage.GetDomains()
}
