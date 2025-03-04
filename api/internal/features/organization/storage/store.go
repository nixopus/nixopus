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

type OrganizationRepository interface {
	GetOrganizations() ([]shared_types.Organization, error)
	CreateOrganization(organization shared_types.Organization) error
	GetOrganization(id string) (*shared_types.Organization, error)
	UpdateOrganization(organization *shared_types.Organization) error
	DeleteOrganization(id string) error
	GetOrganizationUsers(id string) ([]shared_types.OrganizationUsers, error)
	AddUserToOrganization(orgainzation_user shared_types.OrganizationUsers) error
	RemoveUserFromOrganization(user_id string, organization_id string) error
	FindUserInOrganization(user_id string, organization_id string) (*shared_types.OrganizationUsers, error)
	GetOrganizationByName(name string) (*shared_types.Organization, error)
}

// GetOrganizations fetches all organizations from the database.
//
// It uses the bun package to query the database and scan the result into the
// organizations slice.
//
// If the query returns no rows, it returns an empty slice and a nil error.
// If the query returns an error, it returns the error.
func (s OrganizationStore) GetOrganizations() ([]shared_types.Organization, error) {
	var organizations []shared_types.Organization
	err := s.DB.NewSelect().Model(&organizations).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return organizations, nil
	}
	return organizations, err
}

// CreateOrganization creates a new organization in the database.
//
// It takes an organization and inserts it into the database.
//
// If there is an error while creating the organization, it returns the error.
func (s OrganizationStore) CreateOrganization(organization shared_types.Organization) error {
	_, err := s.DB.NewInsert().Model(&organization).Exec(s.Ctx)
	return err
}

// GetOrganization retrieves an organization by its ID from the database.
//
// It queries the database using the provided organization ID and scans the result
// into a shared_types.Organization struct. If no organization with the specified
// ID is found, it returns an empty organization and a nil error.
// If an error occurs during the query, it returns the error.
func (s OrganizationStore) GetOrganization(id string) (*shared_types.Organization, error) {
	organization := &shared_types.Organization{}
	err := s.DB.NewSelect().Model(organization).Where("id = ?", id).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return organization, nil
	}
	return organization, err
}

// UpdateOrganization updates an existing organization in the database.
//
// It takes a pointer to a shared_types.Organization struct and updates the
// corresponding organization in the database.
//
// If there is an error while updating the organization, it returns the error.
func (s OrganizationStore) UpdateOrganization(organization *shared_types.Organization) error {
	_, err := s.DB.NewUpdate().Model(organization).Where("id = ?", organization.ID).Exec(s.Ctx)
	return err
}

// DeleteOrganization deletes an organization from the database by its ID.
//
// It constructs a delete query using the provided organization ID to remove the
// corresponding organization record from the database. If an error occurs during
// the deletion, it returns the error. Otherwise, it returns nil, indicating
// a successful deletion.
func (s OrganizationStore) DeleteOrganization(id string) error {
	_, err := s.DB.NewDelete().Model(&shared_types.Organization{}).Where("id = ?", id).Exec(s.Ctx)
	return err
}

// GetOrganizationUsers retrieves the users for an organization by its ID.
//
// It queries the database using the provided organization ID and scans the result
// into a slice of shared_types.OrganizationUsers structs. The result includes the
// role and user associated with each organization user, as well as the permissions
// associated with each role. If no organization users are found with the specified
// ID, it returns an empty slice and a nil error. If an error occurs during the
// query, it returns the error.
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

// AddUserToOrganization adds a user to an organization.
//
// It takes an organization user and inserts it into the database.
//
// If there is an error while adding the user, it returns the error.
func (s OrganizationStore) AddUserToOrganization(orgainzation_user shared_types.OrganizationUsers) error {
	_, err := s.DB.NewInsert().Model(&orgainzation_user).Exec(s.Ctx)
	return err
}

// GetOrganizationByName retrieves an organization by its name.
//
// It queries the database using the provided name and scans the result
// into a shared_types.Organization struct. If no organization with the
// specified name is found, it returns an empty organization and a nil
// error. If an error occurs during the query, it returns the error.
func (s OrganizationStore) GetOrganizationByName(name string) (*shared_types.Organization, error) {
	organization := &shared_types.Organization{}
	err := s.DB.NewSelect().Model(organization).Where("name = ?", name).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return organization, nil
	}
	return organization, err
}

// FindUserInOrganization retrieves a user from an organization by user ID and organization ID.
//
// It queries the database for an organization user that matches the given user ID and organization ID,
// and checks if the user's record is not marked as deleted. The function returns a pointer to a
// shared_types.OrganizationUsers struct containing the user information if found, and a nil error.
// If no matching user is found, it returns an empty OrganizationUsers struct and a nil error.
// If an error occurs during the query, it returns the error.
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

// RemoveUserFromOrganization removes a user from an organization.
//
// It constructs a delete query that removes the organization user record from the database that
// matches the given user ID and organization ID. If an error occurs during the deletion, it
// returns the error. Otherwise, it returns nil, indicating a successful deletion.
func (s OrganizationStore) RemoveUserFromOrganization(user_id string, organization_id string) error {
	var p shared_types.OrganizationUsers
	_, err := s.DB.NewDelete().Model(&p).Where("user_id = ?", user_id).Where("organization_id = ?", organization_id).Exec(s.Ctx)
	return err
}
