package preferences

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

// newTestDB opens a fresh in-memory SQLite DB and creates the two tables used
// by CheckUserNotificationPreferences. Returns the DB; caller should defer Close.
func newTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=shared&_busy_timeout=5000")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	_, err = db.NewCreateTable().
		Model((*shared_types.NotificationPreferences)(nil)).
		IfNotExists().
		Exec(ctx)
	require.NoError(t, err)

	_, err = db.NewCreateTable().
		Model((*shared_types.PreferenceItem)(nil)).
		IfNotExists().
		Exec(ctx)
	require.NoError(t, err)

	return db
}

// insertPreference inserts a NotificationPreferences row and returns its ID.
func insertPreference(t *testing.T, db *bun.DB, userID uuid.UUID) uuid.UUID {
	t.Helper()
	prefID := uuid.New()
	now := time.Now()
	pref := &shared_types.NotificationPreferences{
		ID:        prefID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := db.NewInsert().Model(pref).Exec(context.Background())
	require.NoError(t, err)
	return prefID
}

// insertPreferenceItem inserts a PreferenceItem row for the given preferenceID.
func insertPreferenceItem(t *testing.T, db *bun.DB, prefID uuid.UUID, category, typ string, enabled bool) {
	t.Helper()
	now := time.Now()
	item := &shared_types.PreferenceItem{
		ID:           uuid.New(),
		PreferenceID: prefID,
		Category:     category,
		Type:         typ,
		Enabled:      enabled,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := db.NewInsert().Model(item).Exec(context.Background())
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// NewPreferenceManager
// ---------------------------------------------------------------------------

func TestNewPreferenceManager_ReturnsNonNil(t *testing.T) {
	db := newTestDB(t)
	m := NewPreferenceManager(db, context.Background())
	require.NotNil(t, m)
}

// ---------------------------------------------------------------------------
// CheckUserNotificationPreferences — invalid userID
// ---------------------------------------------------------------------------

func TestCheckUserNotificationPreferences_InvalidUserID(t *testing.T) {
	db := newTestDB(t)
	m := NewPreferenceManager(db, context.Background())

	ok, err := m.CheckUserNotificationPreferences("not-a-uuid", "security", "login-alerts")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// CheckUserNotificationPreferences — preferences query: sql.ErrNoRows
// ---------------------------------------------------------------------------

func TestCheckUserNotificationPreferences_NoPreferencesRow_ReturnsFalseNil(t *testing.T) {
	db := newTestDB(t)
	m := NewPreferenceManager(db, context.Background())

	// No row inserted for this user → sql.ErrNoRows on the first select.
	ok, err := m.CheckUserNotificationPreferences(uuid.New().String(), "security", "login-alerts")

	require.NoError(t, err)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// CheckUserNotificationPreferences — preferences query: non-ErrNoRows DB error
// ---------------------------------------------------------------------------

func TestCheckUserNotificationPreferences_PreferencesQueryError(t *testing.T) {
	// Open a DB without creating any tables so the first select fails with
	// "no such table: notification_preferences" (not sql.ErrNoRows).
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	m := NewPreferenceManager(db, context.Background())

	ok, err := m.CheckUserNotificationPreferences(uuid.New().String(), "security", "login-alerts")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch user preferences")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// CheckUserNotificationPreferences — preference item query: sql.ErrNoRows
// ---------------------------------------------------------------------------

func TestCheckUserNotificationPreferences_NoPreferenceItem_ReturnsFalseNil(t *testing.T) {
	db := newTestDB(t)
	userID := uuid.New()
	// Preference row exists but no item for this category/type.
	insertPreference(t, db, userID)
	m := NewPreferenceManager(db, context.Background())

	ok, err := m.CheckUserNotificationPreferences(userID.String(), "security", "login-alerts")

	require.NoError(t, err)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// CheckUserNotificationPreferences — preference item query: non-ErrNoRows error
// ---------------------------------------------------------------------------

func TestCheckUserNotificationPreferences_PreferenceItemQueryError(t *testing.T) {
	// Create only the notification_preferences table; omit preference_item so
	// the second select fails with "no such table: preference_item".
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	ctx := context.Background()
	_, err = db.NewCreateTable().
		Model((*shared_types.NotificationPreferences)(nil)).
		IfNotExists().
		Exec(ctx)
	require.NoError(t, err)

	userID := uuid.New()
	insertPreference(t, db, userID)

	m := NewPreferenceManager(db, ctx)

	ok, err := m.CheckUserNotificationPreferences(userID.String(), "security", "login-alerts")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch preference item")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// CheckUserNotificationPreferences — success: enabled = true
// ---------------------------------------------------------------------------

func TestCheckUserNotificationPreferences_EnabledTrue(t *testing.T) {
	db := newTestDB(t)
	userID := uuid.New()
	prefID := insertPreference(t, db, userID)
	insertPreferenceItem(t, db, prefID, "security", "login-alerts", true)
	m := NewPreferenceManager(db, context.Background())

	ok, err := m.CheckUserNotificationPreferences(userID.String(), "security", "login-alerts")

	require.NoError(t, err)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// CheckUserNotificationPreferences — success: enabled = false
// ---------------------------------------------------------------------------

func TestCheckUserNotificationPreferences_EnabledFalse(t *testing.T) {
	db := newTestDB(t)
	userID := uuid.New()
	prefID := insertPreference(t, db, userID)
	insertPreferenceItem(t, db, prefID, "update", "newsletter", false)
	m := NewPreferenceManager(db, context.Background())

	ok, err := m.CheckUserNotificationPreferences(userID.String(), "update", "newsletter")

	require.NoError(t, err)
	assert.False(t, ok)
}
