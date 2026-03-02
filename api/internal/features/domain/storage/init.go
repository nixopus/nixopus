package storage

import (
	"context"

	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type DomainStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

type DomainStorageInterface interface {
	GetDomains(OrganizationID string, UserID uuid.UUID) ([]shared_types.Domain, error)
	CreateCustomDomain(domain *shared_types.Domain) error
	GetCustomDomainsByOrg(orgID uuid.UUID) ([]shared_types.Domain, error)
	GetCustomDomainByID(id uuid.UUID, orgID uuid.UUID) (*shared_types.Domain, error)
	GetCustomDomainByName(name string) (*shared_types.Domain, error)
	UpdateCustomDomainStatus(id uuid.UUID, status string) error
	UpdateCustomDomainVerification(id uuid.UUID, status string, dnsProvider *string) error
	DeleteCustomDomain(id uuid.UUID) error
}

func (s *DomainStorage) getDB() bun.IDB {
	return s.DB
}

func (s *DomainStorage) GetDomains(OrganizationID string, UserID uuid.UUID) ([]shared_types.Domain, error) {
	var domains []shared_types.Domain
	err := s.getDB().NewSelect().Model(&domains).
		Where("organization_id = ? AND deleted_at IS NULL", OrganizationID).
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return domains, nil
}
