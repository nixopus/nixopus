package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type UserStorage struct {
	DB     *bun.DB
	Ctx    context.Context
	Logger *logger.Logger // optional; nil disables storage logs
}

type UserRepository interface {
	GetUserById(id string) (*shared_types.User, error)
	UpdateUserName(userID string, userName string, updatedAt time.Time) error
	GetUserOrganizationsWithRolesAndPermissions(userID string) ([]types.UserOrganizationsResponse, error)
	GetUserSettings(userID string) (*shared_types.UserSettings, error)
	UpdateUserSettings(userID string, updates map[string]interface{}) (*shared_types.UserSettings, error)
	UpdateUserAvatar(ctx context.Context, userID string, avatarData string) error
	GetUserPreferences(userID string) (*shared_types.UserPreferences, error)
	UpdateUserPreferences(userID string, preferences shared_types.UserPreferencesData) (*shared_types.UserPreferences, error)
	GetIsOnboarded(userID string) (bool, error)
	MarkOnboardingComplete(userID string) error
}

func CreateNewUserStorage(db *bun.DB, ctx context.Context, l *logger.Logger) *UserStorage {
	return &UserStorage{
		DB:     db,
		Ctx:    ctx,
		Logger: l,
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
	data := fmt.Sprintf("user_id=%s", id)
	storageLog(s.Logger, logger.Debug, "user storage: GetUserById", data)
	user := &shared_types.User{}
	err := s.DB.NewSelect().Model(user).Where("id = ?", id).Scan(s.Ctx)
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: GetUserById: %v", err), data)
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
	data := fmt.Sprintf("user_id=%s", userID)
	storageLog(s.Logger, logger.Debug, "user storage: UpdateUserName", data)
	_, err := s.DB.NewUpdate().
		Model((*shared_types.User)(nil)).
		Set("name = ?", userName).
		Set("updated_at = ?", updatedAt).
		Where("id = ?", userID).
		Exec(s.Ctx)
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: UpdateUserName: %v", err), data)
	}
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
	data := fmt.Sprintf("user_id=%s", userID)
	storageLog(s.Logger, logger.Debug, "user storage: GetUserOrganizationsWithRolesAndPermissions", data)

	var organizationUsers []shared_types.OrganizationUsers

	query := s.DB.NewSelect().
		TableExpr("member AS ou").
		ColumnExpr("ou.*").
		Join("LEFT JOIN organization AS o ON o.id = ou.organization_id").
		Where("ou.user_id = ?", userID)

	err := query.Scan(s.Ctx, &organizationUsers)
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: GetUserOrganizations member scan: %v", err), data)
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
	data := fmt.Sprintf("user_id=%s", userID)
	storageLog(s.Logger, logger.Debug, "user storage: GetUserSettings", data)

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
				AutoUpdate: false,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			_, err := s.DB.NewInsert().
				Model(defaultSettings).
				Exec(s.Ctx)
			if err != nil {
				storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: GetUserSettings insert defaults: %v", err), data)
				return nil, err
			}
			return defaultSettings, nil
		}
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: GetUserSettings: %v", err), data)
		return nil, err
	}
	return &settings, nil
}

func (s *UserStorage) UpdateUserSettings(userID string, updates map[string]interface{}) (*shared_types.UserSettings, error) {
	data := fmt.Sprintf("user_id=%s", userID)
	storageLog(s.Logger, logger.Debug, "user storage: UpdateUserSettings", data)

	// Ensure a settings row exists before updating.
	if _, err := s.GetUserSettings(userID); err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: UpdateUserSettings ensure row: %v", err), data)
		return nil, err
	}

	query := s.DB.NewUpdate().
		TableExpr("user_settings").
		Where("user_id = ?", userID)

	for key, value := range updates {
		query = query.Set(key+" = ?", value)
	}

	if _, err := query.Exec(s.Ctx); err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: UpdateUserSettings exec: %v", err), data)
		return nil, err
	}

	return s.GetUserSettings(userID)
}

func (s *UserStorage) UpdateUserAvatar(ctx context.Context, userID string, avatarData string) error {
	data := fmt.Sprintf("user_id=%s payload_len=%d", userID, len(avatarData))
	storageLog(s.Logger, logger.Debug, "user storage: UpdateUserAvatar", data)

	var image interface{}
	if avatarData == "" {
		image = nil
	} else {
		image = avatarData
	}
	_, err := s.DB.NewUpdate().
		Model((*shared_types.User)(nil)).
		Set("image = ?", image).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: UpdateUserAvatar: %v", err), data)
		return fmt.Errorf("failed to update user avatar: %w", err)
	}

	return nil
}

// GetUserPreferences retrieves user preferences, creating defaults if none exist
func (s *UserStorage) GetUserPreferences(userID string) (*shared_types.UserPreferences, error) {
	data := fmt.Sprintf("user_id=%s", userID)
	storageLog(s.Logger, logger.Debug, "user storage: GetUserPreferences", data)

	var prefs shared_types.UserPreferences
	err := s.DB.NewSelect().
		Model(&prefs).
		Where("user_id = ?", userID).
		Scan(s.Ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			parsedUserID, parseErr := uuid.Parse(userID)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid user ID: %w", parseErr)
			}

			defaultPrefs := &shared_types.UserPreferences{
				ID:          uuid.New(),
				UserID:      parsedUserID,
				Preferences: shared_types.DefaultUserPreferencesData(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			_, err := s.DB.NewInsert().
				Model(defaultPrefs).
				Exec(s.Ctx)
			if err != nil {
				storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: GetUserPreferences insert defaults: %v", err), data)
				return nil, err
			}
			return defaultPrefs, nil
		}
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: GetUserPreferences: %v", err), data)
		return nil, err
	}
	return &prefs, nil
}

// UpdateUserPreferences updates user preferences with the provided data
func (s *UserStorage) UpdateUserPreferences(userID string, preferences shared_types.UserPreferencesData) (*shared_types.UserPreferences, error) {
	data := fmt.Sprintf("user_id=%s", userID)
	storageLog(s.Logger, logger.Debug, "user storage: UpdateUserPreferences", data)

	// Ensure preferences exist before updating
	existingPrefs, err := s.GetUserPreferences(userID)
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: UpdateUserPreferences load: %v", err), data)
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}
	if existingPrefs == nil {
		return nil, fmt.Errorf("user preferences not found for user ID: %s", userID)
	}

	var prefs shared_types.UserPreferences
	result, err := s.DB.NewUpdate().
		Model(&prefs).
		Set("preferences = ?", preferences).
		Set("updated_at = ?", time.Now()).
		Where("user_id = ?", userID).
		Returning("*").
		Exec(s.Ctx)

	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: UpdateUserPreferences exec: %v", err), data)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: UpdateUserPreferences rows affected: %v", err), data)
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		storageLog(s.Logger, logger.Debug, "user storage: UpdateUserPreferences: no rows updated", data)
		return nil, fmt.Errorf("no rows updated for user ID: %s", userID)
	}

	// Re-fetch to ensure we have the updated data
	updatedPrefs, err := s.GetUserPreferences(userID)
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: UpdateUserPreferences refetch: %v", err), data)
		return nil, fmt.Errorf("failed to refetch user preferences after update: %w", err)
	}
	return updatedPrefs, nil
}

// GetIsOnboarded retrieves the is_onboarded status for a user from the database.
func (s *UserStorage) GetIsOnboarded(userID string) (bool, error) {
	data := fmt.Sprintf("user_id=%s", userID)
	storageLog(s.Logger, logger.Debug, "user storage: GetIsOnboarded", data)

	var user shared_types.User
	err := s.DB.NewSelect().
		Model(&user).
		Column("is_onboarded").
		Where("id = ?", userID).
		Scan(s.Ctx)

	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: GetIsOnboarded: %v", err), data)
		return false, fmt.Errorf("failed to get onboarding status: %w", err)
	}

	return user.IsOnboarded, nil
}

// MarkOnboardingComplete sets the is_onboarded field to true for a user.
func (s *UserStorage) MarkOnboardingComplete(userID string) error {
	data := fmt.Sprintf("user_id=%s", userID)
	storageLog(s.Logger, logger.Debug, "user storage: MarkOnboardingComplete", data)

	_, err := s.DB.NewUpdate().
		Model((*shared_types.User)(nil)).
		Set("is_onboarded = ?", true).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(s.Ctx)

	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("user storage: MarkOnboardingComplete: %v", err), data)
		return fmt.Errorf("failed to mark onboarding as complete: %w", err)
	}

	return nil
}
