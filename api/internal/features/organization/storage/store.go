package storage

import (
	"context"
	"database/sql"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type OrganizationStore struct {
	DB  *bun.DB
	Ctx context.Context
}

func (s OrganizationStore) GetOrganizations() ([]shared_types.Organization, error) {
	var organizations []shared_types.Organization
	err := s.DB.NewSelect().Model(&organizations).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return organizations, nil
	}
	return organizations, err
}

func (s OrganizationStore) CreateOrganization(organization shared_types.Organization) error {
	_, err := s.DB.NewInsert().Model(&organization).Exec(s.Ctx)
	return err
}

func (s OrganizationStore) GetOrganization(id string) (*shared_types.Organization, error) {
	organization := &shared_types.Organization{}
	err := s.DB.NewSelect().Model(organization).Where("id = ?", id).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return organization, nil
	}
	return organization, err
}

func (s OrganizationStore) UpdateOrganization(organization *shared_types.Organization) error {
	_, err := s.DB.NewUpdate().Model(organization).Where("id = ?", organization.ID).Exec(s.Ctx)
	return err
}

func (s OrganizationStore) DeleteOrganization(id string) error {
	_, err := s.DB.NewDelete().Model(&shared_types.Organization{}).Where("id = ?", id).Exec(s.Ctx)
	return err
}

func (s OrganizationStore) GetOrganizationUsers(id string) ([]shared_types.OrganizationUsers, error) {
	var organizationUsers []shared_types.OrganizationUsers
	
	err := s.DB.NewSelect().
		Model(&organizationUsers).
		Where("organization_id = ?", id).
		Relation("Role").
		Relation("User").
		Relation("Role.Permissions").
		Scan(s.Ctx)
	
	if err == sql.ErrNoRows {
		return organizationUsers, nil
	}
	
	return organizationUsers, err
}

func (s OrganizationStore) AddUserToOrganization(orgainzation_user shared_types.OrganizationUsers) error {
	_, err := s.DB.NewInsert().Model(&orgainzation_user).Exec(s.Ctx)
	return err
}

func (s OrganizationStore) GetOrganizationByName(name string) (*shared_types.Organization, error) {
	organization := &shared_types.Organization{}
	err := s.DB.NewSelect().Model(organization).Where("name = ?", name).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return organization, nil
	}
	return organization, err
}

func (s OrganizationStore) FindUserInOrganization(user_id string, organization_id string) (*shared_types.OrganizationUsers, error) {
	organization_user := &shared_types.OrganizationUsers{}
	err := s.DB.NewSelect().
		Model(organization_user).
		Where("user_id = ?", user_id).
		Where("organization_id = ?", organization_id).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(s.Ctx)

	if err == sql.ErrNoRows {
		return organization_user, nil
	}
	return organization_user, err
}

func (s OrganizationStore) RemoveUserFromOrganization(user_id string, organization_id string) error {
	var p shared_types.OrganizationUsers
	_, err := s.DB.NewDelete().Model(&p).Where("user_id = ?", user_id).Where("organization_id = ?", organization_id).Exec(s.Ctx)
	return err
}
