package storage

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/server/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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

	// If status filter is provided, only return SSH keys that have provision details with that status
	if params.Status != "" {
		query = query.Join("INNER JOIN user_provision_details AS upd ON sk.id = upd.ssh_key_id").
			Where("upd.organization_id = ?", orgID).
			Where("upd.status = ?", params.Status)
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

	// If status filter is provided, only count SSH keys that have provision details with that status
	if params.Status != "" {
		countQuery = countQuery.Join("INNER JOIN user_provision_details AS upd ON sk.id = upd.ssh_key_id").
			Where("upd.organization_id = ?", orgID).
			Where("upd.status = ?", params.Status)
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

	// Apply status filter if provided
	if params.Status != "" {
		provisionQuery = provisionQuery.Where("status = ?", params.Status)
	}

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

	// Combine SSH keys with provision details
	servers := make([]types.ServerResponse, 0, len(sshKeys))
	for _, key := range sshKeys {
		server := types.ServerResponse{
			SSHKey: key,
		}
		if provision, ok := provisionMap[key.ID]; ok {
			server.Provision = provision
		}
		servers = append(servers, server)
	}

	return servers, totalCount, nil
}
