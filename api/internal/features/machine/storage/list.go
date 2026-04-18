package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type ListStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

func NewListStorage(db *bun.DB, ctx context.Context) *ListStorage {
	return &ListStorage{DB: db, Ctx: ctx}
}

func (s *ListStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

type ListRepository interface {
	ListMachinesByOrganizationID(orgID uuid.UUID, params types.MachineListParams) ([]types.MachineResponse, int, error)
	SetDefaultMachine(orgID uuid.UUID, machineID uuid.UUID) (*uuid.UUID, error)
	GetMachineByIDAndOrgID(machineID uuid.UUID, orgID uuid.UUID) (*shared_types.SSHKey, error)
}

func (s *ListStorage) ListMachinesByOrganizationID(orgID uuid.UUID, params types.MachineListParams) ([]types.MachineResponse, int, error) {
	query := s.getDB().NewSelect().
		TableExpr("ssh_keys AS sk").
		Where("sk.organization_id = ?", orgID).
		Where("sk.deleted_at IS NULL")

	if params.IsActive != nil {
		query = query.Where("sk.is_active = ?", *params.IsActive)
	}

	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("LOWER(sk.name) LIKE ?", searchPattern).
				WhereOr("LOWER(sk.host) LIKE ?", searchPattern).
				WhereOr("LOWER(COALESCE(sk.description, '')) LIKE ?", searchPattern)
		})
	}

	if params.Status != "" {
		query = query.Join("INNER JOIN user_provision_details AS upd ON sk.id = upd.ssh_key_id").
			Join("INNER JOIN \"user\" AS u ON upd.user_id = u.id").
			Where("upd.organization_id = ?", orgID).
			Where("u.provision_status = ?", params.Status)
	}

	countQuery := s.getDB().NewSelect().
		TableExpr("ssh_keys AS sk").
		Where("sk.organization_id = ?", orgID).
		Where("sk.deleted_at IS NULL")

	if params.IsActive != nil {
		countQuery = countQuery.Where("sk.is_active = ?", *params.IsActive)
	}

	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		countQuery = countQuery.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("LOWER(sk.name) LIKE ?", searchPattern).
				WhereOr("LOWER(sk.host) LIKE ?", searchPattern).
				WhereOr("LOWER(COALESCE(sk.description, '')) LIKE ?", searchPattern)
		})
	}

	if params.Status != "" {
		countQuery = countQuery.Join("INNER JOIN user_provision_details AS upd ON sk.id = upd.ssh_key_id").
			Join("INNER JOIN \"user\" AS u ON upd.user_id = u.id").
			Where("upd.organization_id = ?", orgID).
			Where("u.provision_status = ?", params.Status)
	}

	totalCount, err := countQuery.Count(s.Ctx)
	if err != nil {
		return nil, 0, err
	}

	sortColumn := "sk.created_at"
	if params.SortBy != "" {
		validSortColumns := map[string]string{
			"name":       "sk.name",
			"created_at": "sk.created_at",
			"host":       "sk.host",
			"updated_at": "sk.updated_at",
		}
		if col, ok := validSortColumns[params.SortBy]; ok {
			sortColumn = col
		}
	}

	sortOrder := "DESC"
	if params.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.OrderExpr("? ?", bun.Ident(sortColumn), bun.Safe(sortOrder))

	offset := (params.Page - 1) * params.PageSize
	query = query.Limit(params.PageSize).Offset(offset).ColumnExpr("sk.*")

	var sshKeys []shared_types.SSHKey
	if err = query.Scan(s.Ctx, &sshKeys); err != nil {
		return nil, 0, err
	}

	if len(sshKeys) == 0 {
		return []types.MachineResponse{}, totalCount, nil
	}

	sshKeyIDs := make([]uuid.UUID, 0, len(sshKeys))
	for _, key := range sshKeys {
		sshKeyIDs = append(sshKeyIDs, key.ID)
	}

	var provisionDetails []shared_types.UserProvisionDetails
	if err = s.getDB().NewSelect().
		Model(&provisionDetails).
		Where("ssh_key_id IN (?)", bun.In(sshKeyIDs)).
		Where("organization_id = ?", orgID).
		Scan(s.Ctx); err != nil {
		return nil, 0, err
	}

	provisionMap := make(map[uuid.UUID]*shared_types.UserProvisionDetails)
	for i := range provisionDetails {
		if provisionDetails[i].SSHKeyID != nil {
			provisionMap[*provisionDetails[i].SSHKeyID] = &provisionDetails[i]
		}
	}

	type resourceAggregate struct {
		SSHKeyID    uuid.UUID `bun:"ssh_key_id"`
		TotalVcpu   int       `bun:"total_vcpu"`
		TotalRamMB  int       `bun:"total_ram_mb"`
		TotalDiskGB int       `bun:"total_disk_gb"`
	}
	var aggregates []resourceAggregate
	if err = s.getDB().NewSelect().
		TableExpr("user_provision_details").
		ColumnExpr("ssh_key_id").
		ColumnExpr("COALESCE(SUM(vcpu_count), 0) AS total_vcpu").
		ColumnExpr("COALESCE(SUM(memory_mb), 0) AS total_ram_mb").
		ColumnExpr("COALESCE(SUM(disk_size_gb), 0) AS total_disk_gb").
		Where("ssh_key_id IN (?)", bun.In(sshKeyIDs)).
		Where("organization_id = ?", orgID).
		GroupExpr("ssh_key_id").
		Scan(s.Ctx, &aggregates); err != nil {
		return nil, 0, err
	}

	aggregateMap := make(map[uuid.UUID]resourceAggregate, len(aggregates))
	for _, agg := range aggregates {
		aggregateMap[agg.SSHKeyID] = agg
	}

	machines := make([]types.MachineResponse, 0, len(sshKeys))
	for _, key := range sshKeys {
		m := types.MachineResponse{SSHKey: key}
		if provision, ok := provisionMap[key.ID]; ok {
			m.Provision = provision
		}
		if agg, ok := aggregateMap[key.ID]; ok {
			m.TotalVcpu = agg.TotalVcpu
			m.TotalRamMB = agg.TotalRamMB
			m.TotalDiskGB = agg.TotalDiskGB
		}
		machines = append(machines, m)
	}

	return machines, totalCount, nil
}

func (s *ListStorage) GetMachineByIDAndOrgID(machineID uuid.UUID, orgID uuid.UUID) (*shared_types.SSHKey, error) {
	var key shared_types.SSHKey
	err := s.getDB().NewSelect().
		Model(&key).
		Where("id = ?", machineID).
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *ListStorage) SetDefaultMachine(orgID uuid.UUID, machineID uuid.UUID) (*uuid.UUID, error) {
	tx, err := s.DB.BeginTx(s.Ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var oldKey shared_types.SSHKey
	var oldDefaultID *uuid.UUID
	err = tx.NewSelect().
		Model(&oldKey).
		Column("id").
		Where("organization_id = ?", orgID).
		Where("is_default = ?", true).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		id := oldKey.ID
		oldDefaultID = &id
	}

	var target shared_types.SSHKey
	err = tx.NewSelect().
		Model(&target).
		Column("id", "is_active").
		Where("id = ?", machineID).
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		return nil, types.ErrMachineNotFound
	}
	if !target.IsActive {
		return nil, types.ErrMachineInactive
	}

	_, err = tx.NewUpdate().
		Model((*shared_types.SSHKey)(nil)).
		Set("is_default = ?", false).
		Set("updated_at = NOW()").
		Where("organization_id = ?", orgID).
		Where("is_default = ?", true).
		Exec(s.Ctx)
	if err != nil {
		return nil, err
	}

	_, err = tx.NewUpdate().
		Model((*shared_types.SSHKey)(nil)).
		Set("is_default = ?", true).
		Set("updated_at = NOW()").
		Where("id = ?", machineID).
		Exec(s.Ctx)
	if err != nil {
		return nil, err
	}

	return oldDefaultID, tx.Commit()
}
