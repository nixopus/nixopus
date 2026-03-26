package storage

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type MCPStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

type MCPRepository interface {
	CreateServer(server *shared_types.MCPServer) error
	ListServers(orgID uuid.UUID, enabledOnly bool) ([]shared_types.MCPServer, error)
	GetServerByID(id uuid.UUID) (*shared_types.MCPServer, error)
	GetServerByName(orgID uuid.UUID, name string) (*shared_types.MCPServer, error)
	UpdateServer(server *shared_types.MCPServer) error
	DeleteServer(id uuid.UUID) error
}

func (s MCPStorage) CreateServer(server *shared_types.MCPServer) error {
	_, err := s.DB.NewInsert().Model(server).Exec(s.Ctx)
	return err
}

func (s MCPStorage) ListServers(orgID uuid.UUID, enabledOnly bool) ([]shared_types.MCPServer, error) {
	var servers []shared_types.MCPServer
	q := s.DB.NewSelect().Model(&servers).
		Where("ms.org_id = ? AND ms.deleted_at IS NULL", orgID)
	if enabledOnly {
		q = q.Where("ms.enabled = TRUE")
	}
	err := q.Order("ms.created_at ASC").Scan(s.Ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return []shared_types.MCPServer{}, nil
		}
		return nil, err
	}
	return servers, nil
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
