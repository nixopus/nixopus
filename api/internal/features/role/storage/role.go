package storage

import (
	"context"
	"database/sql"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type RoleStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func NewRoleStorage(db *bun.DB, ctx context.Context) *RoleStorage {
	return &RoleStorage{
		DB:  db,
		Ctx: ctx,
	}
}

// CreateRole creates a new role in the database
//
// It takes in a role object and inserts it into the database.
// It returns an error if the role already exists or if there is a problem
// with the database.
func (s *RoleStorage) CreateRole(role types.Role) error {
	_, err := s.DB.NewInsert().Model(&role).Exec(s.Ctx)
	return err
}

func (s *RoleStorage) GetRoleByName(name string) (*types.Role, error) {
	role := &types.Role{}
	err := s.DB.NewSelect().Model(role).Where("name = ?", name).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return role, nil
	}
	return role, err
}

func (s *RoleStorage) GetRoles() ([]types.Role, error) {
	var roles []types.Role
	err := s.DB.NewSelect().Model(&roles).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return roles, nil
	}
	return roles, err
}

func (s *RoleStorage) GetRole(id string) (*types.Role, error) {
	role := &types.Role{}
	err := s.DB.NewSelect().Model(role).Where("id = ?", id).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return role, nil
	}
	return role, err
}

func (s *RoleStorage) UpdateRole(role *types.Role) error {
	_, err := s.DB.NewUpdate().Model(role).Where("id = ?", role.ID).Exec(s.Ctx)
	return err
}

func (s *RoleStorage) DeleteRole(id string) error {
	_, err := s.DB.NewDelete().Model(&types.Role{}).Where("id = ?", id).Exec(s.Ctx)
	return err
}
