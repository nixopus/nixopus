package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type DomainStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func (s *DomainStorage) CreateDomain(domain *shared_types.Domain) error {
	_, err := s.DB.NewInsert().Model(domain).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DomainStorage) GetDomain(id string) (*shared_types.Domain, error) {
	var domain shared_types.Domain
	err := s.DB.NewSelect().Model(&domain).Where("id = ?", id).Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.ErrDomainNotFound
		}
		return nil, err
	}
	return &domain, nil
}

func (s *DomainStorage) UpdateDomain(ID string,Name string) error {
	var domain shared_types.Domain
	err := s.DB.NewSelect().Model(&domain).Where("id = ?", ID).Scan(s.Ctx)
	if err != nil {
		return err
	}
	domain.Name = Name
	domain.UpdatedAt = time.Now()
	_, err = s.DB.NewUpdate().Model(&domain).Where("id = ?", ID).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DomainStorage) DeleteDomain(domain *shared_types.Domain) error {
	_, err := s.DB.NewDelete().Model(domain).Where("id = ?", domain.ID).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DomainStorage) GetDomains() ([]shared_types.Domain, error) {
	var domains []shared_types.Domain
	err := s.DB.NewSelect().Model(&domains).Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return domains, nil
}

func (s *DomainStorage) GetDomainByName(name string) (*shared_types.Domain, error) {
	var domain shared_types.Domain
	err := s.DB.NewSelect().Model(&domain).Where("name = ?", name).Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil,nil
		}
		return nil, err
	}
	return &domain, nil
}