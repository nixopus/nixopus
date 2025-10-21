package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type UserStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

type UserRepository interface {
	GetUserById(id string) (*shared_types.User, error)
	UpdateUserName(userID string, userName string, updatedAt time.Time) error
	GetUserOrganizationsWithRolesAndPermissions(userID string) ([]types.UserOrganizationsResponse, error)
	GetUserSettings(userID string) (*shared_types.UserSettings, error)
	UpdateUserSettings(userID string, updates map[string]interface{}) (*shared_types.UserSettings, error)
	UpdateUserAvatar(ctx context.Context, userID string, avatarData string) error
}

func CreateNewUserStorage(db *bun.DB, ctx context.Context) *UserStorage {
	return &UserStorage{
		DB:  db,
		Ctx: ctx,
	}
}

// GetUserById retrieves a user by their id from the database.
//
// The function takes a string argument that is the id of the user to be retrieved.
// It queries the database using the bun package and scans the result into a
// shared_types.User struct. If no user with the specified id is found, it returns
// an empty user and a nil error. If an error occurs during the query, it returns
// the error.
func (s *UserStorage) GetUserById(id string) (*shared_types.User, error) {
	user := &shared_types.User{}
	err := s.DB.NewSelect().Model(user).Where("id = ?", id).Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserName updates the username and updated_at fields of a user in the database.
//
// Parameters:
//
//	userID - the unique identifier of the user whose username is to be updated.
//	userName - the new username to set for the user.
//	updatedAt - the timestamp indicating when the update is made.
//
// Returns:
//
//	error - an error if the update query fails, otherwise nil.
func (s *UserStorage) UpdateUserName(userID string, userName string, updatedAt time.Time) error {
	_, err := s.DB.NewUpdate().
		Table("users").
		Set("username = ?", userName).
		Set("updated_at = ?", updatedAt).
		Where("id = ?", userID).
		Exec(s.Ctx)

	return err
}

// GetUserOrganizationsWithRolesAndPermissions retrieves the organizations for a given user.
//
// It first retrieves the organization users for the given user ID, then
// retrieves the associated organization and role for each organization user.
// If an error occurs during the retrieval, it returns the error.
// If the retrieval is successful, it returns a slice of types.UserOrganizationsResponse
// structs containing the organization and role information for each organization user.
func (s *UserStorage) GetUserOrganizationsWithRolesAndPermissions(userID string) ([]types.UserOrganizationsResponse, error) {
	var organizationUsers []shared_types.OrganizationUsers

	query := s.DB.NewSelect().
		TableExpr("organization_users AS ou").
		ColumnExpr("ou.*").
		Join("LEFT JOIN organizations AS o ON o.id = ou.organization_id").
		Where("ou.user_id = ?", userID).
		Where("ou.deleted_at IS NULL")

	err := query.Scan(s.Ctx, &organizationUsers)
	if err != nil {
		return nil, err
	}

	var response []types.UserOrganizationsResponse
	for _, ou := range organizationUsers {
		var organization shared_types.Organization
		err := s.DB.NewSelect().
			Model(&organization).
			Where("id = ?", ou.OrganizationID).
			Scan(s.Ctx)
		if err != nil {
			continue
		}

		orgResponse := types.UserOrganizationsResponse{
			Organization: organization,
		}

		response = append(response, orgResponse)
	}

	return response, nil
}

func (s *UserStorage) GetUserSettings(userID string) (*shared_types.UserSettings, error) {
	var settings shared_types.UserSettings
	err := s.DB.NewSelect().
		Model(&settings).
		Where("user_id = ?", userID).
		Scan(s.Ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			defaultSettings := &shared_types.UserSettings{
				ID:         uuid.New(),
				UserID:     uuid.MustParse(userID),
				FontFamily: "outfit",
				FontSize:   16,
				Language:   "en",
				Theme:      "light",
				AutoUpdate: true,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			_, err := s.DB.NewInsert().
				Model(defaultSettings).
				Exec(s.Ctx)
			if err != nil {
				return nil, err
			}
			return defaultSettings, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (s *UserStorage) UpdateUserSettings(userID string, updates map[string]interface{}) (*shared_types.UserSettings, error) {
	var settings shared_types.UserSettings
	query := s.DB.NewUpdate().
		Model(&settings).
		Where("user_id = ?", userID)

	for key, value := range updates {
		query = query.Set(key+" = ?", value)
	}

	_, err := query.Returning("*").Exec(s.Ctx)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *UserStorage) UpdateUserAvatar(ctx context.Context, userID string, avatarData string) error {
	_, err := s.DB.NewUpdate().
		Table("users").
		Set("avatar = ?", avatarData).
		Set("updated_at = NOW()").
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update user avatar: %w", err)
	}

	return nil
}
