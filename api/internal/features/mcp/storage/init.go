package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type MCPStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

// ListServersParams controls filtering, sorting, and pagination for ListServers.
// Set Limit=0 to return all matching rows (used by the internal agent endpoint).
type ListServersParams struct {
	Q           string
	SortBy      string // "name" | "provider_id" | "created_at"
	SortDir     string // "asc" | "desc"
	Page        int
	Limit       int
	EnabledOnly bool
}

type MCPRepository interface {
	CreateServer(server *shared_types.MCPServer) error
	ListServers(orgID uuid.UUID, params ListServersParams) ([]shared_types.MCPServer, int, error)
	GetServerByID(id uuid.UUID) (*shared_types.MCPServer, error)
	GetServerByName(orgID uuid.UUID, name string) (*shared_types.MCPServer, error)
	UpdateServer(server *shared_types.MCPServer) error
	DeleteServer(id uuid.UUID) error
}

func (s MCPStorage) CreateServer(server *shared_types.MCPServer) error {
	_, err := s.DB.NewInsert().Model(server).Exec(s.Ctx)
	return err
}

func (s MCPStorage) ListServers(orgID uuid.UUID, params ListServersParams) ([]shared_types.MCPServer, int, error) {
	var servers []shared_types.MCPServer

	allowedSort := map[string]bool{"name": true, "created_at": true, "provider_id": true}
	sortBy := "created_at"
	if allowedSort[params.SortBy] {
		sortBy = params.SortBy
	}
	sortDir := "asc"
	if params.SortDir == "desc" {
		sortDir = "desc"
	}

	// Count query (no limit/offset)
	countQ := s.DB.NewSelect().Model((*shared_types.MCPServer)(nil)).
		Where("ms.org_id = ? AND ms.deleted_at IS NULL", orgID)
	if params.EnabledOnly {
		countQ = countQ.Where("ms.enabled = TRUE")
	}
	if params.Q != "" {
		countQ = countQ.Where("ms.name ILIKE ?", "%"+params.Q+"%")
	}
	totalCount, err := countQ.Count(s.Ctx)
	if err != nil {
		return nil, 0, err
	}

	// Data query
	dataQ := s.DB.NewSelect().Model(&servers).
		Where("ms.org_id = ? AND ms.deleted_at IS NULL", orgID)
	if params.EnabledOnly {
		dataQ = dataQ.Where("ms.enabled = TRUE")
	}
	if params.Q != "" {
		dataQ = dataQ.Where("ms.name ILIKE ?", "%"+params.Q+"%")
	}
	dataQ = dataQ.OrderExpr(fmt.Sprintf("ms.%s %s", sortBy, sortDir))

	if params.Limit > 0 {
		offset := (params.Page - 1) * params.Limit
		if offset < 0 {
			offset = 0
		}
		dataQ = dataQ.Limit(params.Limit).Offset(offset)
	}

	err = dataQ.Scan(s.Ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return []shared_types.MCPServer{}, totalCount, nil
		}
		return nil, 0, err
	}
	return servers, totalCount, nil
}

func (s MCPStorage) GetServerByID(id uuid.UUID) (*shared_types.MCPServer, error) {
	server := &shared_types.MCPServer{}
	err := s.DB.NewSelect().Model(server).
		Where("ms.id = ? AND ms.deleted_at IS NULL", id).
		Scan(s.Ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return server, nil
}

func (s MCPStorage) GetServerByName(orgID uuid.UUID, name string) (*shared_types.MCPServer, error) {
	server := &shared_types.MCPServer{}
	err := s.DB.NewSelect().Model(server).
		Where("ms.org_id = ? AND ms.name = ? AND ms.deleted_at IS NULL", orgID, name).
		Scan(s.Ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return server, nil
}

func (s MCPStorage) UpdateServer(server *shared_types.MCPServer) error {
	_, err := s.DB.NewUpdate().Model(server).
		Column("name", "credentials", "custom_url", "enabled", "updated_at").
		Where("id = ? AND deleted_at IS NULL", server.ID).
		Exec(s.Ctx)
	return err
}

func (s MCPStorage) DeleteServer(id uuid.UUID) error {
	_, err := s.DB.NewUpdate().
		Model((*shared_types.MCPServer)(nil)).
		Set("deleted_at = NOW()").
		Set("updated_at = NOW()").
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(s.Ctx)
	return err
}
