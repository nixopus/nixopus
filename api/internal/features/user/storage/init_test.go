package storage_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	user_storage "github.com/nixopus/nixopus/api/internal/features/user/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func newTestUserDB(t *testing.T) (*bun.DB, context.Context) {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file:memuser"+uuid.New().String()+"?mode=memory&cache=shared")
	require.NoError(t, err)

	db := bun.NewDB(sqldb, sqlitedialect.New())
	ctx := context.Background()

	// Register models so bun can resolve M2M relations on User
	db.RegisterModel((*shared_types.Member)(nil))
	db.RegisterModel((*shared_types.OrganizationUsers)(nil))
	db.RegisterModel((*shared_types.User)(nil))
	db.RegisterModel((*shared_types.Organization)(nil))
	db.RegisterModel((*shared_types.UserSettings)(nil))
	db.RegisterModel((*shared_types.UserPreferences)(nil))

	_, err = db.ExecContext(ctx, `CREATE TABLE "user" (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL DEFAULT '',
		email TEXT NOT NULL DEFAULT '',
		email_verified INTEGER NOT NULL DEFAULT 0,
		image TEXT,
		is_onboarded INTEGER NOT NULL DEFAULT 0,
		provision_status TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `CREATE TABLE user_settings (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		font_family TEXT NOT NULL DEFAULT 'system-ui',
		font_size INTEGER NOT NULL DEFAULT 14,
		theme TEXT NOT NULL DEFAULT 'light',
		language TEXT NOT NULL DEFAULT 'en',
		auto_update INTEGER NOT NULL DEFAULT 0,
		deleted_at TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `CREATE TABLE user_preferences (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		preferences TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `CREATE TABLE member (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		organization_id TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'member',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `CREATE TABLE organization (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		logo TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT
	)`)
	require.NoError(t, err)

	t.Cleanup(func() { _ = db.Close() })
	return db, ctx
}

func insertTestUser(t *testing.T, db *bun.DB, ctx context.Context) shared_types.User {
	t.Helper()
	id := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO "user" (id, name, email, is_onboarded, created_at, updated_at) VALUES (?, ?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		id.String(), "testuser", "test@example.com",
	)
	require.NoError(t, err)
	return shared_types.User{ID: id, Name: "testuser", Email: "test@example.com"}
}

// — Constructor —

func TestCreateNewUserStorage(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)
	require.NotNil(t, s)
	assert.Equal(t, db, s.DB)
	assert.Equal(t, ctx, s.Ctx)
}

// — GetUserById —

func TestGetUserById_found(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)
	user := insertTestUser(t, db, ctx)

	result, err := s.GetUserById(user.ID.String())
	require.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Name, result.Name)
}

func TestGetUserById_notFound(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	_, err := s.GetUserById(uuid.New().String())
	require.Error(t, err)
}

func TestGetUserById_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	_, err := s.GetUserById(uuid.New().String())
	require.Error(t, err)
}

// — UpdateUserName —

func TestUpdateUserName_success(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	id := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO "user" (id, name, email, is_onboarded, created_at, updated_at) VALUES (?, ?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		id.String(), "oldname", "patch-name@example.com",
	)
	require.NoError(t, err)

	err = s.UpdateUserName(id.String(), "newname", time.Now())
	require.NoError(t, err)

	var name string
	err = db.NewRaw(`SELECT name FROM "user" WHERE id = ?`, id.String()).Scan(ctx, &name)
	require.NoError(t, err)
	assert.Equal(t, "newname", name)
}

func TestUpdateUserName_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	err := s.UpdateUserName(uuid.New().String(), "name", time.Now())
	require.Error(t, err)
}

// — GetUserOrganizationsWithRolesAndPermissions —

func TestGetUserOrgs_empty(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	orgs, err := s.GetUserOrganizationsWithRolesAndPermissions(uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, orgs)
}

func TestGetUserOrgs_orgNotFound_continue(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)
	user := insertTestUser(t, db, ctx)

	// Member row points to a non-existent organization → hits the "continue" branch
	_, err := db.ExecContext(ctx,
		`INSERT INTO member (id, user_id, organization_id, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		uuid.New().String(), user.ID.String(), uuid.New().String(),
	)
	require.NoError(t, err)

	orgs, err := s.GetUserOrganizationsWithRolesAndPermissions(user.ID.String())
	require.NoError(t, err)
	assert.Empty(t, orgs)
}

func TestGetUserOrgs_withOrgs(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)
	user := insertTestUser(t, db, ctx)

	orgID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO organization (id, name, slug, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		orgID.String(), "Test Org", "test-org",
	)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO member (id, user_id, organization_id, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		uuid.New().String(), user.ID.String(), orgID.String(),
	)
	require.NoError(t, err)

	orgs, err := s.GetUserOrganizationsWithRolesAndPermissions(user.ID.String())
	require.NoError(t, err)
	require.Len(t, orgs, 1)
	assert.Equal(t, orgID, orgs[0].Organization.ID)
}

func TestGetUserOrgs_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	_, err := s.GetUserOrganizationsWithRolesAndPermissions(uuid.New().String())
	require.Error(t, err)
}

// — GetUserSettings —

func TestGetUserSettings_found(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO user_settings (id, user_id, font_family, font_size, theme, language, auto_update) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), userID.String(), "outfit", 16, "dark", "fr", 0,
	)
	require.NoError(t, err)

	settings, err := s.GetUserSettings(userID.String())
	require.NoError(t, err)
	assert.Equal(t, "outfit", settings.FontFamily)
	assert.Equal(t, 16, settings.FontSize)
	assert.Equal(t, "dark", settings.Theme)
}

func TestGetUserSettings_noRows_createsDefault(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	settings, err := s.GetUserSettings(userID.String())
	require.NoError(t, err)
	require.NotNil(t, settings)
	assert.Equal(t, "outfit", settings.FontFamily)
	assert.Equal(t, userID, settings.UserID)
}

func TestGetUserSettings_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	_, err := s.GetUserSettings(uuid.New().String())
	require.Error(t, err)
}

// — UpdateUserSettings —

func TestUpdateUserSettings_success(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO user_settings (id, user_id, font_family, font_size, theme, language, auto_update) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), userID.String(), "system-ui", 14, "light", "en", 0,
	)
	require.NoError(t, err)

	settings, err := s.UpdateUserSettings(userID.String(), map[string]interface{}{
		"font_family": "monospace",
		"font_size":   18,
		"updated_at":  time.Now(),
	})
	require.NoError(t, err)
	require.NotNil(t, settings)
	assert.Equal(t, "monospace", settings.FontFamily)
	assert.Equal(t, 18, settings.FontSize)
}

func TestUpdateUserSettings_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	_, err := s.UpdateUserSettings(uuid.New().String(), map[string]interface{}{"theme": "dark"})
	require.Error(t, err)
}

func TestUpdateUserSettings_createsDefaultsWhenMissing(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	settings, err := s.UpdateUserSettings(userID.String(), map[string]interface{}{
		"theme":      "dark",
		"updated_at": time.Now(),
	})
	require.NoError(t, err)
	require.NotNil(t, settings)
	assert.Equal(t, userID, settings.UserID)
	assert.Equal(t, "dark", settings.Theme)
}

func TestUpdateUserSettings_emptyUpdates_returnsError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO user_settings (id, user_id, font_family, font_size, theme, language, auto_update) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), userID.String(), "system-ui", 14, "light", "en", 0,
	)
	require.NoError(t, err)

	_, err = s.UpdateUserSettings(userID.String(), map[string]interface{}{})
	require.Error(t, err)
}

// — UpdateUserAvatar —

func TestUpdateUserAvatar_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	err := s.UpdateUserAvatar(ctx, uuid.New().String(), "data:image/png;base64,abc")
	require.Error(t, err)
}

func TestUpdateUserAvatar_success(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)
	user := insertTestUser(t, db, ctx)

	err := s.UpdateUserAvatar(ctx, user.ID.String(), "data:image/png;base64,abc")
	require.NoError(t, err)

	var image sql.NullString
	err = db.NewRaw(`SELECT image FROM "user" WHERE id = ?`, user.ID.String()).Scan(ctx, &image)
	require.NoError(t, err)
	require.True(t, image.Valid)
	assert.Equal(t, "data:image/png;base64,abc", image.String)
}

func TestUpdateUserAvatar_clearAvatar(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)
	user := insertTestUser(t, db, ctx)

	_, err := db.ExecContext(ctx, `UPDATE "user" SET image = ? WHERE id = ?`, "existing-avatar", user.ID.String())
	require.NoError(t, err)

	err = s.UpdateUserAvatar(ctx, user.ID.String(), "")
	require.NoError(t, err)

	var image sql.NullString
	err = db.NewRaw(`SELECT image FROM "user" WHERE id = ?`, user.ID.String()).Scan(ctx, &image)
	require.NoError(t, err)
	assert.False(t, image.Valid)
}

// — GetUserPreferences —

func TestGetUserPreferences_found(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO user_preferences (id, user_id, preferences) VALUES (?, ?, ?)`,
		uuid.New().String(), userID.String(), `{"debug_mode":false,"show_api_error_details":false}`,
	)
	require.NoError(t, err)

	prefs, err := s.GetUserPreferences(userID.String())
	require.NoError(t, err)
	require.NotNil(t, prefs)
	assert.Equal(t, userID, prefs.UserID)
}

func TestGetUserPreferences_noRows_createsDefault(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	prefs, err := s.GetUserPreferences(userID.String())
	require.NoError(t, err)
	require.NotNil(t, prefs)
	assert.Equal(t, userID, prefs.UserID)
}

func TestGetUserPreferences_invalidUserID_noRows(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	_, err := s.GetUserPreferences("not-a-valid-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
}

func TestGetUserPreferences_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	_, err := s.GetUserPreferences(uuid.New().String())
	require.Error(t, err)
}

// — UpdateUserPreferences —

func TestUpdateUserPreferences_success(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO user_preferences (id, user_id, preferences) VALUES (?, ?, ?)`,
		uuid.New().String(), userID.String(), `{"debug_mode":false,"show_api_error_details":false}`,
	)
	require.NoError(t, err)

	updated := shared_types.UserPreferencesData{DebugMode: true, ShowApiErrorDetails: true}
	prefs, err := s.UpdateUserPreferences(userID.String(), updated)
	require.NoError(t, err)
	require.NotNil(t, prefs)
	assert.True(t, prefs.Preferences.DebugMode)
}

func TestUpdateUserPreferences_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	_, err := s.UpdateUserPreferences(uuid.New().String(), shared_types.UserPreferencesData{})
	require.Error(t, err)
}

func TestUpdateUserPreferences_createsDefaultsWhenMissing(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	updated := shared_types.UserPreferencesData{
		DebugMode:           true,
		ShowApiErrorDetails: true,
	}

	prefs, err := s.UpdateUserPreferences(userID.String(), updated)
	require.NoError(t, err)
	require.NotNil(t, prefs)
	assert.Equal(t, userID, prefs.UserID)
	assert.True(t, prefs.Preferences.DebugMode)
	assert.True(t, prefs.Preferences.ShowApiErrorDetails)
}

func TestUpdateUserPreferences_noRowsUpdated(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	// Uppercase UUID triggers parse/insert for defaults, but UPDATE ... WHERE user_id = <uppercase>
	// can miss the canonical lowercase row in sqlite text comparison.
	userIDUpper := strings.ToUpper(uuid.New().String())
	_, err := s.UpdateUserPreferences(userIDUpper, shared_types.UserPreferencesData{DebugMode: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no rows updated")
}

// — GetIsOnboarded —

func TestGetIsOnboarded_true(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO "user" (id, name, email, is_onboarded, created_at, updated_at) VALUES (?, ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		userID.String(), "testuser", "test@example.com",
	)
	require.NoError(t, err)

	result, err := s.GetIsOnboarded(userID.String())
	require.NoError(t, err)
	assert.True(t, result)
}

func TestGetIsOnboarded_false(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO "user" (id, name, email, is_onboarded, created_at, updated_at) VALUES (?, ?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		userID.String(), "testuser", "test@example.com",
	)
	require.NoError(t, err)

	result, err := s.GetIsOnboarded(userID.String())
	require.NoError(t, err)
	assert.False(t, result)
}

func TestGetIsOnboarded_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	_, err := s.GetIsOnboarded(uuid.New().String())
	require.Error(t, err)
}

// — MarkOnboardingComplete —

func TestMarkOnboardingComplete_success(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	userID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO "user" (id, name, email, is_onboarded, created_at, updated_at) VALUES (?, ?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		userID.String(), "testuser", "test@example.com",
	)
	require.NoError(t, err)

	err = s.MarkOnboardingComplete(userID.String())
	require.NoError(t, err)

	var isOnboarded bool
	err = db.NewRaw(`SELECT is_onboarded FROM "user" WHERE id = ?`, userID.String()).Scan(ctx, &isOnboarded)
	require.NoError(t, err)
	assert.True(t, isOnboarded)
}

func TestMarkOnboardingComplete_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := newTestUserDB(t)
	require.NoError(t, db.Close())
	s := user_storage.CreateNewUserStorage(db, ctx, nil)

	err := s.MarkOnboardingComplete(uuid.New().String())
	require.Error(t, err)
}
