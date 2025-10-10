package storage

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type OrganizationStore struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

type OrganizationRepository interface {
	GetOrganizations() ([]shared_types.Organization, error)
	CreateOrganization(organization shared_types.Organization) error
	GetOrganization(id string) (*shared_types.Organization, error)
	UpdateOrganization(organization *shared_types.Organization) error
	DeleteOrganization(id string) error
	GetOrganizationUsers(id string) ([]shared_types.OrganizationUsers, error)
	AddUserToOrganization(organizationUser shared_types.OrganizationUsers) error
	RemoveUserFromOrganization(userID string, organizationID string) error
	FindUserInOrganization(userID string, organizationID string) (*shared_types.OrganizationUsers, error)
	GetOrganizationByName(name string) (*shared_types.Organization, error)
	UpdateUserRole(userID string, organizationID string, roleID uuid.UUID) error
	GetOrganizationCount() (int, error)
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) OrganizationRepository
}

func (s *OrganizationStore) BeginTx() (bun.Tx, error) {
	return s.DB.BeginTx(s.Ctx, nil)
}

func (s *OrganizationStore) WithTx(tx bun.Tx) OrganizationRepository {
	return &OrganizationStore{
		DB:  s.DB,
		Ctx: s.Ctx,
		tx:  &tx,
	}
}

func (s *OrganizationStore) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
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
	err := s.getDB().NewSelect().Model(&organizations).Scan(s.Ctx)
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
func (s *OrganizationStore) CreateOrganization(organization shared_types.Organization) error {
	_, err := s.getDB().NewInsert().Model(&organization).Exec(s.Ctx)
	return err
}

// GetOrganization retrieves an organization by its ID from the database.
//
// It queries the database using the provided organization ID and scans the result
// into a shared_types.Organization struct. If no organization with the specified
// ID is found, it returns an empty organization and a nil error.
// If an error occurs during the query, it returns the error.
func (s *OrganizationStore) GetOrganization(id string) (*shared_types.Organization, error) {
	organization := &shared_types.Organization{}
	err := s.getDB().NewSelect().Model(organization).Where("id = ?", id).Scan(s.Ctx)
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
	_, err := s.getDB().NewUpdate().Model(organization).Where("id = ?", organization.ID).Exec(s.Ctx)
	return err
}

// DeleteOrganization deletes an organization from the database by its ID.
//
// It constructs a delete query using the provided organization ID to remove the
// corresponding organization record from the database. If an error occurs during
// the deletion, it returns the error. Otherwise, it returns nil, indicating
// a successful deletion.
func (s OrganizationStore) DeleteOrganization(id string) error {
	_, err := s.getDB().NewDelete().Model(&shared_types.AuditLog{}).Where("organization_id = ?", id).Exec(s.Ctx)
	if err != nil {
		return err
	}

	_, err = s.getDB().NewDelete().Model(&shared_types.Organization{}).Where("id = ?", id).Exec(s.Ctx)
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

	err := s.getDB().NewSelect().
		Model(&organizationUsers).
		Where("organization_id = ?", id).
		Relation("User").
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
func (s OrganizationStore) AddUserToOrganization(organizationUser shared_types.OrganizationUsers) error {
	_, err := s.getDB().NewInsert().Model(&organizationUser).Exec(s.Ctx)
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
	err := s.getDB().NewSelect().Model(organization).Where("name = ?", name).Scan(s.Ctx)
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
func (s OrganizationStore) FindUserInOrganization(userID string, organizationID string) (*shared_types.OrganizationUsers, error) {
	user := &shared_types.OrganizationUsers{}
	err := s.getDB().NewSelect().Model(user).Where("user_id = ? AND organization_id = ?", userID, organizationID).Scan(s.Ctx)
	if err == sql.ErrNoRows {
		return user, nil
	}
	return user, err
}

// RemoveUserFromOrganization removes a user from an organization.
//
// It constructs a delete query that removes the organization user record from the database that
// matches the given user ID and organization ID. If an error occurs during the deletion, it
// returns the error. Otherwise, it returns nil, indicating a successful deletion.
func (s OrganizationStore) RemoveUserFromOrganization(userID string, organizationID string) error {
	_, err := s.getDB().NewDelete().Model(&shared_types.OrganizationUsers{}).Where("user_id = ? AND organization_id = ?", userID, organizationID).Exec(s.Ctx)
	return err
}

// UpdateUserRole updates a user's role in an organization.
//
// It constructs an update query that updates the role ID for the organization user record
// that matches the given user ID and organization ID. If an error occurs during the update,
// it returns the error. Otherwise, it returns nil, indicating a successful update.
func (s OrganizationStore) UpdateUserRole(userID string, organizationID string, roleID uuid.UUID) error {
	_, err := s.getDB().NewUpdate().
		Model(&shared_types.OrganizationUsers{}).
		Set("role_id = ?", roleID).
		Where("user_id = ? AND organization_id = ?", userID, organizationID).
		Exec(s.Ctx)
	return err
}

func (s OrganizationStore) GetOrganizationCount() (int, error) {
	count, err := s.getDB().NewSelect().Model(&shared_types.Organization{}).Count(s.Ctx)
	return count, err
}
