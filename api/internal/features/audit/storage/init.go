package storage

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type AuditStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func NewAuditStorage(db *bun.DB, ctx context.Context) *AuditStorage {
	return &AuditStorage{
		DB:  db,
		Ctx: ctx,
	}
}

func (s *AuditStorage) CreateAuditLog(log *types.AuditLog) error {
	_, err := s.DB.NewInsert().Model(log).Exec(s.Ctx)
	return err
}

func (s *AuditStorage) GetAuditLogs(filters map[string]interface{}, page, pageSize int) ([]*types.AuditLog, int, error) {
	var logs []*types.AuditLog
	query := s.DB.NewSelect().Model(&logs).
		Relation("User").
		Relation("Organization").
		Order("created_at DESC")

	for key, value := range filters {
		query.Where("? = ?", bun.Ident(key), value)
	}

	totalCount, err := query.Count(s.Ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Limit(pageSize).Offset(offset).Scan(s.Ctx)
	if err != nil {
		return nil, 0, err
	}

	return logs, totalCount, nil
}
