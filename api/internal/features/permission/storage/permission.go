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

func (p *PermissionStorage) CreatePermission(permission types.Permission) error {
	_, err := p.DB.NewInsert().Model(&permission).Exec(p.Ctx)
	return err
}

func (p *PermissionStorage) GetPermissions() ([]types.Permission, error) {
	var permissions []types.Permission
	err := p.DB.NewSelect().Model(&permissions).Scan(p.Ctx)
	if err == sql.ErrNoRows {
		return permissions, nil
	}
	return permissions, err
}

func (p *PermissionStorage) GetPermission(id string) (*types.Permission, error) {
	permission := &types.Permission{}
	err := p.DB.NewSelect().Model(permission).Where("id = ?", id).Scan(p.Ctx)
	if err == sql.ErrNoRows {
		return permission, nil
	}
	return permission, err
}

func (p *PermissionStorage) GetPermissionByName(name string) (*types.Permission, error) {
	permission := &types.Permission{}
	err := p.DB.NewSelect().Model(permission).Where("name = ?", name).Scan(p.Ctx)
	if err == sql.ErrNoRows {
		return permission, nil
	}
	return permission, err
}

func (p *PermissionStorage) UpdatePermission(permission *types.Permission) error {
	_, err := p.DB.NewUpdate().Model(permission).Where("id = ?", permission.ID).Exec(p.Ctx)
	return err
}

func (p *PermissionStorage) DeletePermission(id string) error {
	_, err := p.DB.NewDelete().Model(&types.Permission{}).Where("id = ?", id).Exec(p.Ctx)
	return err
}

func (p *PermissionStorage) AddPermissionToRole(permission types.RolePermissions) error {
	_, err := p.DB.NewInsert().Model(&permission).Exec(p.Ctx)
	return err
}

func (p *PermissionStorage) RemovePermissionFromRole(permission_id string) error {
	var rp types.RolePermissions
	_, err := p.DB.NewDelete().Model(&rp).Where("permission_id = ?", permission_id).Exec(p.Ctx)
	return err
}

func (p *PermissionStorage) GetPermissionsByRole(id string) ([]types.RolePermissions, error) {
	var permissions []types.RolePermissions
	err := p.DB.NewSelect().Model(&permissions).Where("role_id = ?", id).Scan(p.Ctx)
	if err == sql.ErrNoRows {
		return permissions, nil
	}
	return permissions, err
}
