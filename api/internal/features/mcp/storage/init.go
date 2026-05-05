package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type MCPStorage struct {
	DB     *bun.DB
	Ctx    context.Context
	Logger *logger.Logger // optional; nil disables storage logs
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
	ctxStr := fmt.Sprintf("org_id=%s server_id=%s", server.OrgID, server.ID)
	storageLog(s.Logger, logger.Debug, "storage: CreateServer", ctxStr)
	_, err := s.DB.NewInsert().Model(server).Exec(s.Ctx)
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("storage: CreateServer: %v", err), ctxStr)
	}
	return err
}

func (s MCPStorage) ListServers(orgID uuid.UUID, params ListServersParams) ([]shared_types.MCPServer, int, error) {
	ctxStr := fmt.Sprintf("org_id=%s", orgID)
	storageLog(s.Logger, logger.Debug, "storage: ListServers", ctxStr)

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
		storageLog(s.Logger, logger.Error, fmt.Sprintf("storage: ListServers count: %v", err), ctxStr)
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
		if errors.Is(err, sql.ErrNoRows) {
			storageLog(s.Logger, logger.Debug, "storage: ListServers no rows", ctxStr)
			return []shared_types.MCPServer{}, totalCount, nil
		}
		storageLog(s.Logger, logger.Error, fmt.Sprintf("storage: ListServers scan: %v", err), ctxStr)
		return nil, 0, err
	}
	return servers, totalCount, nil
}

func (s MCPStorage) GetServerByID(id uuid.UUID) (*shared_types.MCPServer, error) {
	ctxStr := fmt.Sprintf("server_id=%s", id)
	storageLog(s.Logger, logger.Debug, "storage: GetServerByID", ctxStr)

	server := &shared_types.MCPServer{}
	err := s.DB.NewSelect().Model(server).
		Where("ms.id = ? AND ms.deleted_at IS NULL", id).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			storageLog(s.Logger, logger.Debug, "storage: GetServerByID not found", ctxStr)
			return nil, nil
		}
		storageLog(s.Logger, logger.Error, fmt.Sprintf("storage: GetServerByID: %v", err), ctxStr)
		return nil, err
	}
	return server, nil
}

func (s MCPStorage) GetServerByName(orgID uuid.UUID, name string) (*shared_types.MCPServer, error) {
	ctxStr := fmt.Sprintf("org_id=%s name=%s", orgID, name)
	storageLog(s.Logger, logger.Debug, "storage: GetServerByName", ctxStr)

	server := &shared_types.MCPServer{}
	err := s.DB.NewSelect().Model(server).
		Where("ms.org_id = ? AND ms.name = ? AND ms.deleted_at IS NULL", orgID, name).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			storageLog(s.Logger, logger.Debug, "storage: GetServerByName not found", ctxStr)
			return nil, nil
		}
		storageLog(s.Logger, logger.Error, fmt.Sprintf("storage: GetServerByName: %v", err), ctxStr)
		return nil, err
	}
	return server, nil
}

func (s MCPStorage) UpdateServer(server *shared_types.MCPServer) error {
	ctxStr := fmt.Sprintf("server_id=%s org_id=%s", server.ID, server.OrgID)
	storageLog(s.Logger, logger.Debug, "storage: UpdateServer", ctxStr)
	_, err := s.DB.NewUpdate().Model(server).
		Column("name", "credentials", "custom_url", "enabled", "updated_at").
		Where("id = ? AND deleted_at IS NULL", server.ID).
		Exec(s.Ctx)
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("storage: UpdateServer: %v", err), ctxStr)
	}
	return err
}

func (s MCPStorage) DeleteServer(id uuid.UUID) error {
	ctxStr := fmt.Sprintf("server_id=%s", id)
	storageLog(s.Logger, logger.Debug, "storage: DeleteServer (soft)", ctxStr)
	_, err := s.DB.NewUpdate().
		Model((*shared_types.MCPServer)(nil)).
		Set("deleted_at = NOW()").
		Set("updated_at = NOW()").
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(s.Ctx)
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("storage: DeleteServer: %v", err), ctxStr)
	}
	return err
}
