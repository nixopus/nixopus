package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/uptrace/bun"
)

type BackupStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func NewBackupStorage(db *bun.DB, ctx context.Context) *BackupStorage {
	return &BackupStorage{DB: db, Ctx: ctx}
}

func (s *BackupStorage) ListByOrg(ctx context.Context, orgID uuid.UUID, params types.BackupListParams) ([]types.MachineBackup, int, error) {
	query := s.DB.NewSelect().
		Model((*types.MachineBackup)(nil)).
		Where("mb.organization_id = ?", orgID)

	countQuery := s.DB.NewSelect().
		Model((*types.MachineBackup)(nil)).
		Where("mb.organization_id = ?", orgID)

	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		applySearch := func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
				return sq.Where("LOWER(mb.machine_name) LIKE ?", searchPattern).
					WhereOr("LOWER(COALESCE(mb.error, '')) LIKE ?", searchPattern)
			})
		}
		query = applySearch(query)
		countQuery = applySearch(countQuery)
	}

	if params.Status != "" {
		query = query.Where("mb.status = ?", params.Status)
		countQuery = countQuery.Where("mb.status = ?", params.Status)
	}

	totalCount, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count backups: %w", err)
	}

	sortColumn := "mb.created_at"
	validSortColumns := map[string]string{
		"created_at": "mb.created_at",
		"status":     "mb.status",
		"size_bytes": "mb.size_bytes",
	}
	if col, ok := validSortColumns[params.SortBy]; ok {
		sortColumn = col
	}

	sortOrder := "DESC"
	if params.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.OrderExpr("? ?", bun.Ident(sortColumn), bun.Safe(sortOrder))

	offset := (params.Page - 1) * params.PageSize
	query = query.Limit(params.PageSize).Offset(offset)

	var backups []types.MachineBackup
	err = query.Scan(ctx, &backups)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list backups: %w", err)
	}
	return backups, totalCount, nil
}

func (s *BackupStorage) HasInProgressBackup(ctx context.Context, orgID uuid.UUID) (bool, error) {
	exists, err := s.DB.NewSelect().
		Model((*types.MachineBackup)(nil)).
		Where("organization_id = ?", orgID).
		Where("status IN (?)", bun.In([]types.MachineBackupStatus{
			types.BackupStatusPending,
			types.BackupStatusInProgress,
		})).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check in-progress backups: %w", err)
	}
	return exists, nil
}

func (s *BackupStorage) GetLatestCompletedBackup(ctx context.Context, orgID uuid.UUID) (*types.MachineBackup, error) {
	var backup types.MachineBackup
	err := s.DB.NewSelect().
		Model(&backup).
		Where("organization_id = ?", orgID).
		Where("status = ?", types.BackupStatusCompleted).
		OrderExpr("completed_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest completed backup: %w", err)
	}
	return &backup, nil
}

func (s *BackupStorage) GetByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*types.MachineBackup, error) {
	var backup types.MachineBackup
	err := s.DB.NewSelect().
		Model(&backup).
		Where("id = ? AND organization_id = ?", id, orgID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}
	return &backup, nil
}
