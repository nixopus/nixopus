package storage

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *DomainStorage) CreateCustomDomain(domain *shared_types.Domain) error {
	domain.Type = "custom"
	_, err := s.getDB().NewInsert().Model(domain).Exec(s.Ctx)
	return err
}

func (s *DomainStorage) GetCustomDomainsByOrg(orgID uuid.UUID) ([]shared_types.Domain, error) {
	var domains []shared_types.Domain
	err := s.getDB().NewSelect().Model(&domains).
		Where("type = ?", "custom").
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return domains, nil
}

func (s *DomainStorage) GetCustomDomainByID(id uuid.UUID, orgID uuid.UUID) (*shared_types.Domain, error) {
	var domain shared_types.Domain
	err := s.getDB().NewSelect().Model(&domain).
		Where("id = ?", id).
		Where("organization_id = ?", orgID).
		Where("type = ?", "custom").
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.ErrCustomDomainNotFound
		}
		return nil, err
	}
	return &domain, nil
}

func (s *DomainStorage) GetCustomDomainByName(name string) (*shared_types.Domain, error) {
	var domain shared_types.Domain
	err := s.getDB().NewSelect().Model(&domain).
		Where("name = ?", name).
		Where("type = ?", "custom").
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain, nil
}

func (s *DomainStorage) UpdateCustomDomainStatus(id uuid.UUID, status string) error {
	_, err := s.getDB().NewUpdate().
		Model((*shared_types.Domain)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(s.Ctx)
	return err
}

func (s *DomainStorage) UpdateCustomDomainVerification(id uuid.UUID, status string, dnsProvider *string) error {
	_, err := s.getDB().NewUpdate().
		Model((*shared_types.Domain)(nil)).
		Set("status = ?", status).
		Set("dns_provider = ?", dnsProvider).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(s.Ctx)
	return err
}

func (s *DomainStorage) DeleteCustomDomain(id uuid.UUID) error {
	now := time.Now()
	result, err := s.getDB().NewUpdate().
		Model((*shared_types.Domain)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(s.Ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return types.ErrCustomDomainNotFound
	}

	return nil
}
