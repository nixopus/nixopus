package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func (s *BackupStorage) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]types.MachineBackup, error) {
	var backups []types.MachineBackup
	err := s.DB.NewSelect().
		Model(&backups).
		Where("organization_id = ?", orgID).
		OrderExpr("created_at DESC").
		Limit(50).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}
	return backups, nil
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
