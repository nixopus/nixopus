package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type DomainStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

type DomainStorageInterface interface {
	CreateDomain(domain *shared_types.Domain) error
	GetDomain(id string) (*shared_types.Domain, error)
	UpdateDomain(ID string, Name string) error
	DeleteDomain(domain *shared_types.Domain) error
	GetDomains(OrganizationID string, UserID uuid.UUID) ([]shared_types.Domain, error)
	GetDomainByName(name string, organizationID uuid.UUID) (*shared_types.Domain, error)
	IsDomainExists(ID string) (bool, error)
	GetDomainOwnerByID(ID string) (string, error)
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) DomainStorageInterface
}

func (s *DomainStorage) BeginTx() (bun.Tx, error) {
	return s.DB.BeginTx(s.Ctx, nil)
}

func (s *DomainStorage) WithTx(tx bun.Tx) DomainStorageInterface {
	return &DomainStorage{
		DB:  s.DB,
		Ctx: s.Ctx,
		tx:  &tx,
	}
}

func (s *DomainStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

func (s *DomainStorage) CreateDomain(domain *shared_types.Domain) error {
	_, err := s.getDB().NewInsert().Model(domain).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DomainStorage) GetDomain(id string) (*shared_types.Domain, error) {
	var domain shared_types.Domain
	err := s.getDB().NewSelect().Model(&domain).Where("id = ? AND deleted_at IS NULL", id).Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.ErrDomainNotFound
		}
		return nil, err
	}
	return &domain, nil
}

func (s *DomainStorage) UpdateDomain(ID string, Name string) error {
	var domain shared_types.Domain
	err := s.getDB().NewSelect().Model(&domain).Where("id = ? AND deleted_at IS NULL", ID).Scan(s.Ctx)
	if err != nil {
		return err
	}
	domain.Name = Name
	domain.UpdatedAt = time.Now()
	_, err = s.getDB().NewUpdate().Model(&domain).Where("id = ? AND deleted_at IS NULL", ID).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DomainStorage) DeleteDomain(domain *shared_types.Domain) error {
	now := time.Now()
	result, err := s.getDB().NewUpdate().Model(domain).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ? AND deleted_at IS NULL", domain.ID).
		Exec(s.Ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return types.ErrDomainNotFound
	}

	return nil
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

func (s *DomainStorage) GetDomainByName(name string, organizationID uuid.UUID) (*shared_types.Domain, error) {
	var domain shared_types.Domain
	err := s.getDB().NewSelect().Model(&domain).
		Where("name = ? AND organization_id = ? AND deleted_at IS NULL", name, organizationID).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain, nil
}

func (s *DomainStorage) IsDomainExists(ID string) (bool, error) {
	var domain shared_types.Domain
	err := s.getDB().NewSelect().Model(&domain).Where("id = ? AND deleted_at IS NULL", ID).Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *DomainStorage) GetDomainOwnerByID(ID string) (string, error) {
	var domain shared_types.Domain
	err := s.getDB().NewSelect().Model(&domain).Where("id = ? AND deleted_at IS NULL", ID).Scan(s.Ctx)
	if err != nil {
		return "", err
	}
	return domain.UserID.String(), nil
}
