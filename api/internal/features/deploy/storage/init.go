package storage

import (
	"context"
	"github.com/uptrace/bun"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type DeployStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

type DeployRepository interface {
	IsNameAlreadyTaken(name string) (bool, error)
	IsDomainAlreadyTaken(domain string) (bool, error)
	IsPortAlreadyTaken(port int) (bool, error)
	IsDomainValid(domain string) (bool, error)
	AddApplication(application *shared_types.Application) error
	AddApplicationLogs(applicationLogs *shared_types.ApplicationLogs) error
	AddApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error
	GetApplications(page int, pageSize int) ([]shared_types.Application, int, error)
	UpdateApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error
}

func (s *DeployStorage) IsNameAlreadyTaken(name string) (bool, error) {
	var count int
	err := s.DB.NewSelect().
		TableExpr("applications").
		ColumnExpr("count(*)").
		Where("name = ?", name).
		Scan(s.Ctx, &count)

	return count > 0, err
}

func (s *DeployStorage) IsDomainAlreadyTaken(domain string) (bool, error) {
	var count int
	err := s.DB.NewSelect().
		TableExpr("applications").
		ColumnExpr("count(*)").
		Where("domain = ?", domain).
		Scan(s.Ctx, &count)

	return count > 0, err
}

func (s *DeployStorage) IsPortAlreadyTaken(port int) (bool, error) {
	var count int
	err := s.DB.NewSelect().
		TableExpr("applications").
		ColumnExpr("count(*)").
		Where("port = ?", port).
		Scan(s.Ctx, &count)

	return count > 0, err
}

func (s *DeployStorage) IsDomainValid(domain string) (bool, error) {
	return true, nil
}

func (s *DeployStorage) AddApplication(application *shared_types.Application) error {
	_, err := s.DB.NewInsert().Model(application).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) AddApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error {
	_, err := s.DB.NewInsert().Model(applicationStatus).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) UpdateApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error {
	_, err := s.DB.NewUpdate().Model(applicationStatus).Where("id = ?", applicationStatus.ID).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) AddApplicationLogs(applicationLogs *shared_types.ApplicationLogs) error {
	_, err := s.DB.NewInsert().Model(applicationLogs).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) GetApplications(page, pageSize int) ([]shared_types.Application, int, error) {
	var applications []shared_types.Application

	offset := (page - 1) * pageSize

	totalCount, err := s.DB.NewSelect().
		Model((*shared_types.Application)(nil)).
		Count(s.Ctx)

	if err != nil {
		return nil, 0, err
	}

	err = s.DB.NewSelect().
		Model(&applications).
		Relation("Domain").
		Relation("User").
		Relation("Status").
		Relation("Logs").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(s.Ctx)

	if err != nil {
		return nil, 0, err
	}

	return applications, totalCount, nil
}
