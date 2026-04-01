package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/server/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type ServerStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

func (s *ServerStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

// ServerRepository defines the interface for server storage operations.
// This enables mocking in tests.
type ServerRepository interface {
	ListServersByOrganizationID(orgID uuid.UUID, params types.ServerListParams) ([]types.ServerResponse, int, error)
	SetDefaultServer(orgID uuid.UUID, serverID uuid.UUID) (*uuid.UUID, error)
	GetServerByIDAndOrgID(serverID uuid.UUID, orgID uuid.UUID) (*shared_types.SSHKey, error)
}

// ListServersByOrganizationID retrieves all SSH keys (servers) for an organization
// with optional user_provision_details joined as provision field.
// Returns servers, total count, and error.
func (s *ServerStorage) ListServersByOrganizationID(orgID uuid.UUID, params types.ServerListParams) ([]types.ServerResponse, int, error) {
	// Build base query for SSH keys
	query := s.getDB().NewSelect().
		TableExpr("ssh_keys AS sk").
		Where("sk.organization_id = ?", orgID).
		Where("sk.deleted_at IS NULL")

	// Apply is_active filter if provided
	if params.IsActive != nil {
		query = query.Where("sk.is_active = ?", *params.IsActive)
	}

	// Apply search filter
	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("LOWER(sk.name) LIKE ?", searchPattern).
				WhereOr("LOWER(sk.host) LIKE ?", searchPattern).
				WhereOr("LOWER(COALESCE(sk.description, '')) LIKE ?", searchPattern)
		})
	}

	// If status filter is provided, filter by user.provision_status
	if params.Status != "" {
		query = query.Join("INNER JOIN user_provision_details AS upd ON sk.id = upd.ssh_key_id").
			Join("INNER JOIN \"user\" AS u ON upd.user_id = u.id").
			Where("upd.organization_id = ?", orgID).
			Where("u.provision_status = ?", params.Status)
	}

	// Build count query for total count
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

	// If status filter is provided, filter count by user.provision_status
	if params.Status != "" {
		countQuery = countQuery.Join("INNER JOIN user_provision_details AS upd ON sk.id = upd.ssh_key_id").
			Join("INNER JOIN \"user\" AS u ON upd.user_id = u.id").
			Where("upd.organization_id = ?", orgID).
			Where("u.provision_status = ?", params.Status)
	}

	// Get total count
	totalCount, err := countQuery.Count(s.Ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortColumn := "sk.created_at"
	if params.SortBy != "" {
		// Validate sort column to prevent SQL injection
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

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	query = query.Limit(params.PageSize).Offset(offset)

	// Select SSH key columns
	query = query.ColumnExpr("sk.*")

	// Execute query to get SSH keys
	var sshKeys []shared_types.SSHKey
	err = query.Scan(s.Ctx, &sshKeys)
	if err != nil {
		return nil, 0, err
	}

	if len(sshKeys) == 0 {
		return []types.ServerResponse{}, totalCount, nil
	}

	// Get SSH key IDs
	sshKeyIDs := make([]uuid.UUID, 0, len(sshKeys))
	for _, key := range sshKeys {
		sshKeyIDs = append(sshKeyIDs, key.ID)
	}

	// Query user_provision_details for these SSH keys
	var provisionDetails []shared_types.UserProvisionDetails
	provisionQuery := s.getDB().NewSelect().
		Model(&provisionDetails).
		Where("ssh_key_id IN (?)", bun.In(sshKeyIDs)).
		Where("organization_id = ?", orgID)

	err = provisionQuery.Scan(s.Ctx)
	if err != nil {
		return nil, 0, err
	}

	// Create a map of provision details by SSH key ID
	provisionMap := make(map[uuid.UUID]*shared_types.UserProvisionDetails)
	for i := range provisionDetails {
		if provisionDetails[i].SSHKeyID != nil {
			provisionMap[*provisionDetails[i].SSHKeyID] = &provisionDetails[i]
		}
	}

	// Aggregate resource totals (vcpu, ram, disk) per SSH key
	type resourceAggregate struct {
		SSHKeyID    uuid.UUID `bun:"ssh_key_id"`
		TotalVcpu   int       `bun:"total_vcpu"`
		TotalRamMB  int       `bun:"total_ram_mb"`
		TotalDiskGB int       `bun:"total_disk_gb"`
	}
	var aggregates []resourceAggregate
	err = s.getDB().NewSelect().
		TableExpr("user_provision_details").
		ColumnExpr("ssh_key_id").
		ColumnExpr("COALESCE(SUM(vcpu_count), 0) AS total_vcpu").
		ColumnExpr("COALESCE(SUM(memory_mb), 0) AS total_ram_mb").
		ColumnExpr("COALESCE(SUM(disk_size_gb), 0) AS total_disk_gb").
		Where("ssh_key_id IN (?)", bun.In(sshKeyIDs)).
		Where("organization_id = ?", orgID).
		GroupExpr("ssh_key_id").
		Scan(s.Ctx, &aggregates)
	if err != nil {
		return nil, 0, err
	}

	aggregateMap := make(map[uuid.UUID]resourceAggregate, len(aggregates))
	for _, agg := range aggregates {
		aggregateMap[agg.SSHKeyID] = agg
	}

	// Combine SSH keys with provision details and resource totals
	servers := make([]types.ServerResponse, 0, len(sshKeys))
	for _, key := range sshKeys {
		server := types.ServerResponse{
			SSHKey: key,
		}
		if provision, ok := provisionMap[key.ID]; ok {
			server.Provision = provision
		}
		if agg, ok := aggregateMap[key.ID]; ok {
			server.TotalVcpu = agg.TotalVcpu
			server.TotalRamMB = agg.TotalRamMB
			server.TotalDiskGB = agg.TotalDiskGB
		}
		servers = append(servers, server)
	}

	return servers, totalCount, nil
}

// GetServerByIDAndOrgID retrieves a non-deleted SSH key that belongs to the given org.
// Returns sql.ErrNoRows if not found or wrong org.
func (s *ServerStorage) GetServerByIDAndOrgID(serverID uuid.UUID, orgID uuid.UUID) (*shared_types.SSHKey, error) {
	var key shared_types.SSHKey
	err := s.getDB().NewSelect().
		Model(&key).
		Where("id = ?", serverID).
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// SetDefaultServer atomically designates serverID as the org's default server.
// Returns the previous default's ID (nil if none existed).
// Returns types.ErrServerNotFound if target not found/wrong org, types.ErrServerInactive if inactive.
func (s *ServerStorage) SetDefaultServer(orgID uuid.UUID, serverID uuid.UUID) (*uuid.UUID, error) {
	tx, err := s.DB.BeginTx(s.Ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Step 1: capture old default ID
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

	// Step 2: validate target server (must exist, be in org, and be active)
	var target shared_types.SSHKey
	err = tx.NewSelect().
		Model(&target).
		Column("id", "is_active").
		Where("id = ?", serverID).
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		return nil, types.ErrServerNotFound
	}
	if !target.IsActive {
		return nil, types.ErrServerInactive
	}

	// Step 3: clear existing default
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

	// Step 4: set new default
	_, err = tx.NewUpdate().
		Model((*shared_types.SSHKey)(nil)).
		Set("is_default = ?", true).
		Set("updated_at = NOW()").
		Where("id = ?", serverID).
		Exec(s.Ctx)
	if err != nil {
		return nil, err
	}

	return oldDefaultID, tx.Commit()
}
