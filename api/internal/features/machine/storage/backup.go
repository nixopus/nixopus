package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

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

func (s *BackupStorage) ListByOrg(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, params types.BackupListParams) ([]types.MachineBackup, int, error) {
	query := s.DB.NewSelect().
		Model((*types.MachineBackup)(nil)).
		Where("mb.organization_id = ?", orgID)

	countQuery := s.DB.NewSelect().
		Model((*types.MachineBackup)(nil)).
		Where("mb.organization_id = ?", orgID)

	if serverID != nil {
		serverFilter := func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Join("JOIN user_provision_details upd ON upd.id = mb.provision_id").
				Where("(upd.ssh_key_id = ? OR upd.server_id = ?)", *serverID, *serverID)
		}
		query = serverFilter(query)
		countQuery = serverFilter(countQuery)
	}

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

func (s *BackupStorage) HasInProgressBackup(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (bool, error) {
	q := s.DB.NewSelect().
		Model((*types.MachineBackup)(nil)).
		Where("mb.organization_id = ?", orgID).
		Where("mb.status IN (?)", bun.In([]types.MachineBackupStatus{
			types.BackupStatusPending,
			types.BackupStatusInProgress,
		}))

	if serverID != nil {
		q = q.Join("JOIN user_provision_details upd ON upd.id = mb.provision_id").
			Where("(upd.ssh_key_id = ? OR upd.server_id = ?)", *serverID, *serverID)
	}

	exists, err := q.Exists(ctx)
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

func (s *BackupStorage) InsertBackup(ctx context.Context, backup *types.MachineBackup) error {
	_, err := s.DB.NewInsert().
		Model(backup).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert backup: %w", err)
	}
	return nil
}

func (s *BackupStorage) UpdateBackupStatus(ctx context.Context, id uuid.UUID, status types.MachineBackupStatus, updates map[string]interface{}) error {
	q := s.DB.NewUpdate().
		TableExpr("machine_backups").
		Where("id = ?", id).
		Set("status = ?", status).
		Set("updated_at = ?", time.Now())

	for k, v := range updates {
		q = q.Set(k+" = ?", v)
	}

	_, err := q.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update backup status: %w", err)
	}
	return nil
}

// GetProvisionIDBySSHKey returns the user_provision_details.id for a BYOS machine,
// matched by ssh_key_id. Returns nil (no error) if no row is found.
func (s *BackupStorage) GetProvisionIDBySSHKey(ctx context.Context, orgID, sshKeyID uuid.UUID) (*uuid.UUID, error) {
	var row struct {
		ID uuid.UUID `bun:"id"`
	}
	err := s.DB.NewSelect().
		TableExpr("user_provision_details").
		ColumnExpr("id").
		Where("ssh_key_id = ?", sshKeyID).
		Where("organization_id = ?", orgID).
		Where("type = 'user_owned'").
		Limit(1).
		Scan(ctx, &row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get provision ID: %w", err)
	}
	return &row.ID, nil
}
