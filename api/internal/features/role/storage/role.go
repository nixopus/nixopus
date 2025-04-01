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
	tx  *bun.Tx
}

func NewRoleStorage(db *bun.DB, ctx context.Context) *RoleStorage {
	return &RoleStorage{
		DB:  db,
		Ctx: ctx,
	}
}

type RoleRepository interface {
	CreateRole(role types.Role) error
	GetRoleByName(name string) (*types.Role, error)
	GetRoles() ([]types.Role, error)
	GetRole(id string) (*types.Role, error)
	UpdateRole(role *types.Role) error
	DeleteRole(id string) error
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) RoleRepository
}

func (s *RoleStorage) BeginTx() (bun.Tx, error) {
	return s.DB.BeginTx(s.Ctx, nil)
}

func (s *RoleStorage) WithTx(tx bun.Tx) RoleRepository {
	return &RoleStorage{
		DB:  s.DB,
		Ctx: s.Ctx,
		tx:  &tx,
	}
}

func (s *RoleStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

// CreateRole creates a new role in the database
//
// It takes in a role object and inserts it into the database.
// It returns an error if the role already exists or if there is a problem
// with the database.
func (s *RoleStorage) CreateRole(role types.Role) error {
	_, err := s.getDB().NewInsert().Model(&role).Exec(s.Ctx)
	return err
}

// GetRoleByName retrieves a role by its name from the storage.
// It returns the role and nil if found, or an error if the operation fails.
// If the role does not exist, it returns a nil role and no error.
func (s *RoleStorage) GetRoleByName(name string) (*types.Role, error) {
	role := &types.Role{}
	err := s.getDB().NewSelect().Model(role).Where("name = ?", name).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return role, nil
	}
	return role, err
}

// GetRoles retrieves all roles from the storage.
// It returns a slice of roles and nil if the operation is successful, or an error if the operation fails.
// If no roles are found, it returns an empty slice and no error.
func (s *RoleStorage) GetRoles() ([]types.Role, error) {
	var roles []types.Role
	err := s.getDB().NewSelect().
		Model(&roles).
		Relation("Permissions").
		Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return roles, nil
	}
	return roles, err
}

// GetRole retrieves a role by its ID from the storage.
// It returns the role and nil if found, or an error if the operation fails.
// If the role does not exist, it returns a nil role and no error.
func (s *RoleStorage) GetRole(id string) (*types.Role, error) {
	role := &types.Role{}
	err := s.getDB().NewSelect().Model(role).Where("id = ?", id).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return role, nil
	}
	return role, err
}

// UpdateRole updates an existing role in the database.
//
// It takes a pointer to a Role object, which contains the updated
// role information. The update is performed based on the role's ID.
// If the update operation is successful, it returns nil. Otherwise,
// it returns an error indicating what went wrong.
func (s *RoleStorage) UpdateRole(role *types.Role) error {
	_, err := s.getDB().NewUpdate().Model(role).Where("id = ?", role.ID).Exec(s.Ctx)
	return err
}

// DeleteRole deletes a role from the database by its ID.
//
// It constructs a delete query using the provided role ID to remove the
// corresponding role record from the database. If an error occurs during
// the deletion, it returns the error. Otherwise, it returns nil, indicating
// a successful deletion.
func (s *RoleStorage) DeleteRole(id string) error {
	_, err := s.getDB().NewDelete().Model(&types.Role{}).Where("id = ?", id).Exec(s.Ctx)
	return err
}
