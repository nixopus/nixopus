package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type ServerStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

type ServerStorageInterface interface {
	CreateServer(server *shared_types.Server) error
	GetServer(id string) (*shared_types.Server, error)
	UpdateServer(server *shared_types.Server) error
	UpdateServerStatus(ID string, Status string) error
	DeleteServer(server *shared_types.Server) error
	GetServers(OrganizationID string, UserID uuid.UUID) ([]shared_types.Server, error)
	GetServersPaginated(OrganizationID string, UserID uuid.UUID, queryParams *types.ServerQueryParams) ([]shared_types.Server, error)
	GetServersCount(OrganizationID string, UserID uuid.UUID, search string) (int, error)
	GetServerName(name string, organizationID uuid.UUID) (*shared_types.Server, error)
	GetServerByHost(host string, port int, organizationID uuid.UUID) (*shared_types.Server, error)
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) ServerStorageInterface
}

func (s *ServerStorage) BeginTx() (bun.Tx, error) {
	return s.DB.BeginTx(s.Ctx, nil)
}

func (s *ServerStorage) WithTx(tx bun.Tx) ServerStorageInterface {
	return &ServerStorage{
		DB:  s.DB,
		Ctx: s.Ctx,
		tx:  &tx,
	}
}

func (s *ServerStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

func (s *ServerStorage) CreateServer(server *shared_types.Server) error {
	_, err := s.getDB().NewInsert().Model(server).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServerStorage) GetServer(id string) (*shared_types.Server, error) {
	var server shared_types.Server
	err := s.getDB().NewSelect().Model(&server).Where("id = ? AND deleted_at IS NULL", id).Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *ServerStorage) UpdateServer(server *shared_types.Server) error {
	_, err := s.getDB().NewUpdate().Model(server).
		Column("name", "description", "host", "port", "username", "ssh_password", "ssh_private_key_path", "updated_at").
		Where("id = ? AND deleted_at IS NULL", server.ID).
		Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServerStorage) UpdateServerStatus(ID string, Status string) error {
	_, err := s.getDB().NewUpdate().
		Model((*shared_types.Server)(nil)).
		Set("status = ?, updated_at = NOW()", Status).
		Where("id = ? AND deleted_at IS NULL", ID).
		Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServerStorage) DeleteServer(server *shared_types.Server) error {
	_, err := s.getDB().NewUpdate().Model(server).
		Set("deleted_at = ?", "NOW()").
		Where("id = ? AND deleted_at IS NULL", server.ID).
		Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServerStorage) GetServers(OrganizationID string, UserID uuid.UUID) ([]shared_types.Server, error) {
	var servers []shared_types.Server
	err := s.getDB().NewSelect().Model(&servers).
		Column("id", "name", "description", "host", "port", "username", "status", "created_at", "updated_at", "user_id", "organization_id").
		Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL", OrganizationID, UserID).
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *ServerStorage) GetServersPaginated(OrganizationID string, UserID uuid.UUID, queryParams *types.ServerQueryParams) ([]shared_types.Server, error) {
	var servers []shared_types.Server

	query := s.getDB().NewSelect().Model(&servers).
		Column("id", "name", "description", "host", "port", "username", "status", "created_at", "updated_at", "user_id", "organization_id").
		Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL", OrganizationID, UserID)

	if queryParams.Search != "" {
		searchTerm := "%" + queryParams.Search + "%"
		query = query.Where("(name ILIKE ? OR host ILIKE ? OR username ILIKE ? OR description ILIKE ?)",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	sortField := queryParams.SortBy
	sortOrder := queryParams.SortOrder

	switch sortField {
	case "created_at", "updated_at", "name", "host", "port", "username":
	default:
		sortField = "created_at"
		sortOrder = "desc"
	}

	query = query.Order(sortField + " " + sortOrder)

	query = query.Limit(queryParams.GetLimit()).Offset(queryParams.GetOffset())

	err := query.Scan(s.Ctx)
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func (s *ServerStorage) GetServersCount(OrganizationID string, UserID uuid.UUID, search string) (int, error) {
	query := s.getDB().NewSelect().Model((*shared_types.Server)(nil)).
		Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL", OrganizationID, UserID)

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("(name ILIKE ? OR host ILIKE ? OR username ILIKE ? OR description ILIKE ?)",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	count, err := query.Count(s.Ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *ServerStorage) GetServerName(name string, organizationID uuid.UUID) (*shared_types.Server, error) {
	var server shared_types.Server
	err := s.getDB().NewSelect().Model(&server).
		Where("name = ? AND organization_id = ? AND deleted_at IS NULL", name, organizationID).
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *ServerStorage) GetServerByHost(host string, port int, organizationID uuid.UUID) (*shared_types.Server, error) {
	var server shared_types.Server
	err := s.getDB().NewSelect().Model(&server).
		Where("host = ? AND port = ? AND organization_id = ? AND deleted_at IS NULL", host, port, organizationID).
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return &server, nil
}
