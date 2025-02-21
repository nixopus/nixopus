package storage

import (
	"context"
	"database/sql"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

// CreateRole creates a new role in the database
//
// It takes in a role object and inserts it into the database.
// It returns an error if the role already exists or if there is a problem
// with the database.
func CreateRole(db *bun.DB, role types.Role, ctx context.Context) error {
	_, err := db.NewInsert().Model(&role).Exec(ctx)
	return err
}

func GetRoleByName(db *bun.DB, name string, ctx context.Context) (*types.Role, error) {
	role := &types.Role{}
	err := db.NewSelect().Model(role).Where("name = ?", name).Scan(ctx)
	if err == sql.ErrNoRows {
		return role, nil
	}
	return role, err
}

func GetRoles(db *bun.DB, ctx context.Context) ([]types.Role, error) {
	var roles []types.Role
	err := db.NewSelect().Model(&roles).Scan(ctx)
	if err == sql.ErrNoRows {
		return roles, nil
	}
	return roles, err
}

func GetRole(db *bun.DB, id string, ctx context.Context) (*types.Role, error) {
	role := &types.Role{}
	err := db.NewSelect().Model(role).Where("id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return role, nil
	}
	return role, err
}

func UpdateRole(db *bun.DB, role *types.Role, ctx context.Context) error {
	_, err := db.NewUpdate().Model(role).Where("id = ?", role.ID).Exec(ctx)
	return err
}

func DeleteRole(db *bun.DB, id string, ctx context.Context) error {
	_, err := db.NewDelete().Model(&types.Role{}).Where("id = ?", id).Exec(ctx)
	return err
}

func CreatePermission(db *bun.DB, permission types.Permission, ctx context.Context) error {
	_, err := db.NewInsert().Model(&permission).Exec(ctx)
	return err
}

func GetPermissions(db *bun.DB, ctx context.Context) ([]types.Permission, error) {
	var permissions []types.Permission
	err := db.NewSelect().Model(&permissions).Scan(ctx)
	if err == sql.ErrNoRows {
		return permissions, nil
	}
	return permissions, err
}

func GetPermission(db *bun.DB, id string, ctx context.Context) (*types.Permission, error) {
	permission := &types.Permission{}
	err := db.NewSelect().Model(permission).Where("id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return permission, nil
	}
	return permission, err
}

func GetPermissionByName(db *bun.DB, name string, ctx context.Context) (*types.Permission, error) {
	permission := &types.Permission{}
	err := db.NewSelect().Model(permission).Where("name = ?", name).Scan(ctx)
	if err == sql.ErrNoRows {
		return permission, nil
	}
	return permission, err
}

func UpdatePermission(db *bun.DB, permission *types.Permission, ctx context.Context) error {
	_, err := db.NewUpdate().Model(permission).Where("id = ?", permission.ID).Exec(ctx)
	return err
}

func DeletePermission(db *bun.DB, id string, ctx context.Context) error {
	_, err := db.NewDelete().Model(&types.Permission{}).Where("id = ?", id).Exec(ctx)
	return err
}

func AddPermissionToRole(db *bun.DB, permission types.RolePermissions, ctx context.Context) error {
	_, err := db.NewInsert().Model(&permission).Exec(ctx)
	return err
}

func RemovePermissionFromRole(db *bun.DB, permission_id string, ctx context.Context) error {
	var p types.RolePermissions
	_, err := db.NewDelete().Model(&p).Where("permission_id = ?", permission_id).Exec(ctx)
	return err
}

func GetPermissionsByRole(db *bun.DB, id string, ctx context.Context) ([]types.RolePermissions, error) {
	var permissions []types.RolePermissions
	err := db.NewSelect().Model(&permissions).Where("role_id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return permissions, nil
	}
	return permissions, err
}

func GetOrganizations(db *bun.DB, ctx context.Context) ([]types.Organization, error) {
	var organizations []types.Organization
	err := db.NewSelect().Model(&organizations).Scan(ctx)
	if err == sql.ErrNoRows {
		return organizations, nil
	}
	return organizations, err
}

func CreateOrganization(db *bun.DB, organization types.Organization, ctx context.Context) error {
	_, err := db.NewInsert().Model(&organization).Exec(ctx)
	return err
}

func GetOrganization(db *bun.DB, id string, ctx context.Context) (*types.Organization, error) {
	organization := &types.Organization{}
	err := db.NewSelect().Model(organization).Where("id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return organization, nil
	}
	return organization, err
}

func UpdateOrganization(db *bun.DB, organization *types.Organization, ctx context.Context) error {
	_, err := db.NewUpdate().Model(organization).Where("id = ?", organization.ID).Exec(ctx)
	return err
}

func DeleteOrganization(db *bun.DB, id string, ctx context.Context) error {
	_, err := db.NewDelete().Model(&types.Organization{}).Where("id = ?", id).Exec(ctx)
	return err
}

func GetOrganizationUsers(db *bun.DB, id string, ctx context.Context) ([]types.OrganizationUsers, error) {
	var organization_users []types.OrganizationUsers
	err := db.NewSelect().Model(&organization_users).Where("organization_id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return organization_users, nil
	}
	return organization_users, err
}

func AddUserToOrganization(db *bun.DB, orgainzation_user types.OrganizationUsers, ctx context.Context) error {
	_, err := db.NewInsert().Model(&orgainzation_user).Exec(ctx)
	return err
}

func GetOrganizationByName(db *bun.DB, name string, ctx context.Context) (*types.Organization, error) {
	organization := &types.Organization{}
	err := db.NewSelect().Model(organization).Where("name = ?", name).Scan(ctx)
	if err == sql.ErrNoRows {
		return organization, nil
	}
	return organization, err
}

func FindUserInOrganization(db *bun.DB, user_id string, organization_id string, ctx context.Context) (*types.OrganizationUsers, error) {
	organization_user := &types.OrganizationUsers{}
	err := db.NewSelect().
		Model(organization_user).
		Where("user_id = ?", user_id).
		Where("organization_id = ?", organization_id).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return organization_user, nil
	}
	return organization_user, err
}

func RemoveUserFromOrganization(db *bun.DB, user_id string, organization_id string, ctx context.Context) error {
	var p types.OrganizationUsers
	_, err := db.NewDelete().Model(&p).Where("user_id = ?", user_id).Where("organization_id = ?", organization_id).Exec(ctx)
	return err
}