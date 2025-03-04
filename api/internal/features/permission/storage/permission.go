package storage

import (
	"context"
	"database/sql"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type PermissionStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

type PermissionRepository interface {
	CreatePermission(permission types.Permission) error
	GetPermissions() ([]types.Permission, error)
	GetPermission(id string) (*types.Permission, error)
	GetPermissionByNameAndResource(name string, resource string) (*types.Permission, error)
	UpdatePermission(permission *types.Permission) error
	DeletePermission(id string) error
	AddPermissionToRole(permission types.RolePermissions) error
	RemovePermissionFromRole(permission_id string) error
	GetPermissionsByRole(role_id string) ([]types.RolePermissions, error)
}

// CreatePermission creates a new permission in the application.
//
// It takes a Permission object, which it uses to create a new entry in the
// permissions table.
// If the creation fails, it returns an error.
func (p *PermissionStorage) CreatePermission(permission types.Permission) error {
	_, err := p.DB.NewInsert().Model(&permission).Exec(p.Ctx)
	return err
}

// GetPermissions retrieves all permissions from the application.
//
// It returns a slice of Permission objects or an error if retrieval fails.
func (p *PermissionStorage) GetPermissions() ([]types.Permission, error) {
	var permissions []types.Permission
	err := p.DB.NewSelect().Model(&permissions).Scan(p.Ctx)
	if err == sql.ErrNoRows {
		return permissions, nil
	}
	return permissions, err
}

// GetPermission retrieves a permission by its ID.
//
// It returns a Permission object or an error if retrieval fails.
// If the permission does not exist, it returns a nil Permission and no error.
func (p *PermissionStorage) GetPermission(id string) (*types.Permission, error) {
	permission := &types.Permission{}
	err := p.DB.NewSelect().Model(permission).Where("id = ?", id).Scan(p.Ctx)
	if err == sql.ErrNoRows {
		return permission, nil
	}
	return permission, err
}

// GetPermissionByNameAndResource retrieves a permission by its name and resource.
//
// It returns a Permission object or an error if retrieval fails.
// If the permission does not exist, it returns a nil Permission and no error.
func (p *PermissionStorage) GetPermissionByNameAndResource(name string, resource string) (*types.Permission, error) {
	permission := &types.Permission{}
	err := p.DB.NewSelect().Model(permission).Where("name = ?", name).Where("resource = ?", resource).Scan(p.Ctx)
	if err == sql.ErrNoRows {
		return permission, nil
	}
	return permission, err
}

// UpdatePermission updates a permission in the application.
//
// It takes a Permission object, which it uses to update the corresponding entry in the
// permissions table.
// If the update fails, it returns an error.
func (p *PermissionStorage) UpdatePermission(permission *types.Permission) error {
	_, err := p.DB.NewUpdate().Model(permission).Where("id = ?", permission.ID).Exec(p.Ctx)
	return err
}

// DeletePermission deletes a permission from the application.
//
// It takes an ID, which it uses to delete the corresponding entry in the
// permissions table.
// If the deletion fails, it returns an error.
func (p *PermissionStorage) DeletePermission(id string) error {
	_, err := p.DB.NewDelete().Model(&types.Permission{}).Where("id = ?", id).Exec(p.Ctx)
	return err
}

// AddPermissionToRole adds a permission to a role by creating a new
// RolePermissions entry associating the permission with the role.
// If the addition fails, it returns an error.
func (p *PermissionStorage) AddPermissionToRole(permission types.RolePermissions) error {
	_, err := p.DB.NewInsert().Model(&permission).Exec(p.Ctx)
	return err
}

// RemovePermissionFromRole removes a permission from a role by the permission ID.
//
// It removes the entry in the role_permissions table that associates the
// permission with the role.
// If the removal fails, it returns an error.
func (p *PermissionStorage) RemovePermissionFromRole(permission_id string) error {
	var rp types.RolePermissions
	_, err := p.DB.NewDelete().Model(&rp).Where("permission_id = ?", permission_id).Exec(p.Ctx)
	return err
}

// GetPermissionsByRole retrieves the permissions associated with a specific role ID.
//
// It returns a slice of RolePermissions objects or an error if retrieval fails.
// If the role does not exist, it returns a nil slice and no error.
func (p *PermissionStorage) GetPermissionsByRole(id string) ([]types.RolePermissions, error) {
	var permissions []types.RolePermissions
	err := p.DB.NewSelect().Model(&permissions).Where("role_id = ?", id).Scan(p.Ctx)
	if err == sql.ErrNoRows {
		return permissions, nil
	}
	return permissions, err
}
