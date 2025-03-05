package storage

import (
	"context"
	"github.com/uptrace/bun"
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
