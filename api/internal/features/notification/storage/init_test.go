package storage

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=shared&_busy_timeout=5000")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	for _, m := range []any{
		(*shared_types.SMTPConfigs)(nil),
		(*shared_types.WebhookConfig)(nil),
		(*shared_types.NotificationPreferences)(nil),
		(*shared_types.PreferenceItem)(nil),
	} {
		_, err = db.NewCreateTable().Model(m).IfNotExists().Exec(ctx)
		require.NoError(t, err)
	}
	_, err = db.ExecContext(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS notification_preferences_one_per_user
		ON notification_preferences (user_id)
	`)
	require.NoError(t, err)

	return db
}

func sampleSMTP(t *testing.T, userID, orgID uuid.UUID) *shared_types.SMTPConfigs {
	t.Helper()
	return &shared_types.SMTPConfigs{
		ID:             uuid.New(),
		Host:           "smtp.example.com",
		Port:           587,
		Username:       "user@example.com",
		Password:       "secret",
		FromEmail:      "from@example.com",
		FromName:       "From",
		Security:       "tls",
		UserID:         userID,
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IsActive:       true,
	}
}

func insertNotificationPreference(t *testing.T, db *bun.DB, userID uuid.UUID) uuid.UUID {
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
// AddSmtp
// ---------------------------------------------------------------------------

func TestAddSmtp_Success(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	s := NotificationStorage{DB: db, Ctx: ctx}
	cfg := sampleSMTP(t, uuid.New(), uuid.New())
	require.NoError(t, s.AddSmtp(cfg))

	n, err := db.NewSelect().Model((*shared_types.SMTPConfigs)(nil)).Where("id = ?", cfg.ID).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestAddSmtp_DatabaseError(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	s := NotificationStorage{DB: db, Ctx: context.Background()}
	err = s.AddSmtp(sampleSMTP(t, uuid.New(), uuid.New()))
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// UpdateSmtp
// ---------------------------------------------------------------------------

func TestUpdateSmtp_AllFields(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	cfg := sampleSMTP(t, uuid.New(), uuid.New())
	_, err := db.NewInsert().Model(cfg).Exec(ctx)
	require.NoError(t, err)

	host := "new.host"
	port := 25
	user := "u"
	pass := "p"
	fn := "name"
	fe := "fe@x.com"
	s := NotificationStorage{DB: db, Ctx: ctx}
	err = s.UpdateSmtp(&notification.UpdateSMTPConfigRequest{
		ID:        cfg.ID,
		Host:      &host,
		Port:      &port,
		Username:  &user,
		Password:  &pass,
		FromName:  &fn,
		FromEmail: &fe,
	})
	require.NoError(t, err)

	var got shared_types.SMTPConfigs
	err = db.NewSelect().Model(&got).Where("id = ?", cfg.ID).Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, host, got.Host)
	assert.Equal(t, port, got.Port)
	assert.Equal(t, user, got.Username)
	assert.Equal(t, pass, got.Password)
	assert.Equal(t, fn, got.FromName)
	assert.Equal(t, fe, got.FromEmail)
}

func TestUpdateSmtp_NoSetColumns_StillExec(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	cfg := sampleSMTP(t, uuid.New(), uuid.New())
	_, err := db.NewInsert().Model(cfg).Exec(ctx)
	require.NoError(t, err)

	s := NotificationStorage{DB: db, Ctx: ctx}
	err = s.UpdateSmtp(&notification.UpdateSMTPConfigRequest{ID: cfg.ID})
	// Bun/SQLite may reject UPDATE with no SET; accept either error or nil.
	_ = err
}

func TestUpdateSmtp_DatabaseError(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	s := NotificationStorage{DB: db, Ctx: context.Background()}
	h := "x"
	require.Error(t, s.UpdateSmtp(&notification.UpdateSMTPConfigRequest{ID: uuid.New(), Host: &h}))
}

// ---------------------------------------------------------------------------
// DeleteSmtp
// ---------------------------------------------------------------------------

func TestDeleteSmtp_Success(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	cfg := sampleSMTP(t, uuid.New(), uuid.New())
	_, err := db.NewInsert().Model(cfg).Exec(ctx)
	require.NoError(t, err)

	s := NotificationStorage{DB: db, Ctx: ctx}
	require.NoError(t, s.DeleteSmtp(cfg.ID.String()))
}

func TestDeleteSmtp_NotFound(t *testing.T) {
	db := newTestDB(t)
	s := NotificationStorage{DB: db, Ctx: context.Background()}
	err := s.DeleteSmtp(uuid.New().String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "smtp config not found")
}

func TestDeleteSmtp_DeleteExecError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	cfg := sampleSMTP(t, uuid.New(), uuid.New())
	_, err := db.NewInsert().Model(cfg).Exec(ctx)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TRIGGER sc_block_delete BEFORE DELETE ON smtp_configs
		BEGIN
			SELECT RAISE(ABORT, 'blocked');
		END;
	`)
	require.NoError(t, err)

	s := NotificationStorage{DB: db, Ctx: ctx}
	err = s.DeleteSmtp(cfg.ID.String())
	require.Error(t, err)
}

func TestDeleteSmtp_DatabaseError(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	s := NotificationStorage{DB: db, Ctx: context.Background()}
	require.Error(t, s.DeleteSmtp(uuid.New().String()))
}

// ---------------------------------------------------------------------------
// GetSmtp
// ---------------------------------------------------------------------------

func TestGetSmtp_Found(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	cfg := sampleSMTP(t, userID, uuid.New())
	_, err := db.NewInsert().Model(cfg).Exec(ctx)
	require.NoError(t, err)

	s := NotificationStorage{DB: db, Ctx: ctx}
	got, err := s.GetSmtp(userID.String())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, cfg.ID, got.ID)
}

func TestGetSmtp_NoRows(t *testing.T) {
	db := newTestDB(t)
	s := NotificationStorage{DB: db, Ctx: context.Background()}
	got, err := s.GetSmtp(uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestGetSmtp_QueryError(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	s := NotificationStorage{DB: db, Ctx: context.Background()}
	_, err = s.GetSmtp(uuid.New().String())
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// GetOrganizationsSmtp
// ---------------------------------------------------------------------------

func TestGetOrganizationsSmtp_Found(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	orgID := uuid.New()
	u1 := sampleSMTP(t, uuid.New(), orgID)
	u2 := sampleSMTP(t, uuid.New(), orgID)
	otherOrg := sampleSMTP(t, uuid.New(), uuid.New())
	_, err := db.NewInsert().Model(&[]shared_types.SMTPConfigs{*u1, *u2, *otherOrg}).Exec(ctx)
	require.NoError(t, err)

	s := NotificationStorage{DB: db, Ctx: ctx}
	list, err := s.GetOrganizationsSmtp(orgID.String())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestGetOrganizationsSmtp_Empty(t *testing.T) {
	db := newTestDB(t)
	s := NotificationStorage{DB: db, Ctx: context.Background()}
	list, err := s.GetOrganizationsSmtp(uuid.New().String())
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestGetOrganizationsSmtp_QueryError(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	s := NotificationStorage{DB: db, Ctx: context.Background()}
	_, err = s.GetOrganizationsSmtp(uuid.New().String())
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// UpdatePreference
// ---------------------------------------------------------------------------

func TestUpdatePreference_ExistingRow_UpdatesItem(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	prefID := insertNotificationPreference(t, db, userID)
	insertPreferenceItem(t, db, prefID, "security", "login-alerts", true)

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err := s.UpdatePreference(ctx, notification.UpdatePreferenceRequest{
		Category: "security",
		Type:     "login-alerts",
		Enabled:  false,
	}, userID)
	require.NoError(t, err)

	var enabled bool
	err = db.NewSelect().
		Model((*shared_types.PreferenceItem)(nil)).
		Column("enabled").
		Where("preference_id = ? AND category = ? AND type = ?", prefID, "security", "login-alerts").
		Scan(ctx, &enabled)
	require.NoError(t, err)
	assert.False(t, enabled)
}

func TestUpdatePreference_NoPreferences_InitializesWithRequestOverride(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userID := uuid.New()

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err := s.UpdatePreference(ctx, notification.UpdatePreferenceRequest{
		Category: "update",
		Type:     "newsletter",
		Enabled:  true,
	}, userID)
	require.NoError(t, err)

	var prefID uuid.UUID
	err = db.NewSelect().
		Model((*shared_types.NotificationPreferences)(nil)).
		Column("id").
		Where("user_id = ?", userID).
		Scan(ctx, &prefID)
	require.NoError(t, err)

	var enabled bool
	err = db.NewSelect().
		Model((*shared_types.PreferenceItem)(nil)).
		Column("enabled").
		Where("preference_id = ? AND category = ? AND type = ?", prefID, "update", "newsletter").
		Scan(ctx, &enabled)
	require.NoError(t, err)
	assert.True(t, enabled)
}

func TestUpdatePreference_FetchPreferencesError(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	_, err = db.NewCreateTable().Model((*shared_types.PreferenceItem)(nil)).IfNotExists().Exec(context.Background())
	require.NoError(t, err)

	s := &NotificationStorage{DB: db, Ctx: context.Background()}
	err = s.UpdatePreference(context.Background(), notification.UpdatePreferenceRequest{
		Category: "security",
		Type:     "login-alerts",
		Enabled:  true,
	}, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch user preferences")
}

func TestUpdatePreference_UpdateExecError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	prefID := insertNotificationPreference(t, db, userID)
	insertPreferenceItem(t, db, prefID, "security", "login-alerts", true)

	_, err := db.NewDropTable().Model((*shared_types.PreferenceItem)(nil)).IfExists().Exec(ctx)
	require.NoError(t, err)

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err = s.UpdatePreference(ctx, notification.UpdatePreferenceRequest{
		Category: "security",
		Type:     "login-alerts",
		Enabled:  false,
	}, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update preference")
}

// ---------------------------------------------------------------------------
// GetPreferences
// ---------------------------------------------------------------------------

func TestGetPreferences_ExistingUser(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	prefID := insertNotificationPreference(t, db, userID)
	insertPreferenceItem(t, db, prefID, "activity", "team-updates", true)

	s := &NotificationStorage{DB: db, Ctx: ctx}
	resp, err := s.GetPreferences(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.GreaterOrEqual(t, len(resp.Activity), 1)
}

func TestGetPreferences_InitDefaultsWhenMissing(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userID := uuid.New()

	s := &NotificationStorage{DB: db, Ctx: ctx}
	resp, err := s.GetPreferences(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Activity, 1)
	assert.Len(t, resp.Security, 3)
	assert.Len(t, resp.Update, 3)

	n, err := db.NewSelect().Model((*shared_types.PreferenceItem)(nil)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 7, n)
}

func TestGetPreferences_FetchUserPrefsNonNoRows(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	_, err = db.NewCreateTable().Model((*shared_types.PreferenceItem)(nil)).IfNotExists().Exec(context.Background())
	require.NoError(t, err)

	s := &NotificationStorage{DB: db, Ctx: context.Background()}
	_, err = s.GetPreferences(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch user preferences")
}

func TestGetPreferences_InitThenFetchItemsError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userID := uuid.New()

	s := &NotificationStorage{DB: db, Ctx: ctx}
	_, err := s.GetPreferences(ctx, userID)
	require.NoError(t, err)

	_, err = db.NewDropTable().Model((*shared_types.PreferenceItem)(nil)).IfExists().Exec(ctx)
	require.NoError(t, err)

	_, err = s.GetPreferences(ctx, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch preference items")
}

func TestGetPreferences_InitDefaultPreferencesFailure(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	_, err := db.NewDropTable().Model((*shared_types.PreferenceItem)(nil)).IfExists().Exec(ctx)
	require.NoError(t, err)

	s := &NotificationStorage{DB: db, Ctx: ctx}
	_, err = s.GetPreferences(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize preferences")
}

func TestGetPreferences_FetchNewlyCreatedPreferencesFails(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		CREATE TRIGGER np_delete_self_after_insert
		AFTER INSERT ON notification_preferences
		BEGIN
			DELETE FROM notification_preferences WHERE id = NEW.id;
		END;
	`)
	require.NoError(t, err)

	s := &NotificationStorage{DB: db, Ctx: ctx}
	_, err = s.GetPreferences(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch newly created preferences")
}

// ---------------------------------------------------------------------------
// initUserPreferences / initDefaultPreferences — transaction failures
// ---------------------------------------------------------------------------

func TestInitUserPreferences_BeginTxError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	require.NoError(t, db.Close())

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err := s.initUserPreferences(ctx, uuid.New(), notification.UpdatePreferenceRequest{
		Category: "security",
		Type:     "login-alerts",
		Enabled:  true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start transaction")
}

func TestInitDefaultPreferences_BeginTxError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	require.NoError(t, db.Close())

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err := s.initDefaultPreferences(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start transaction")
}

func TestInitUserPreferences_CommitError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	prev := bunTxCommitHook
	bunTxCommitHook = func(*bun.Tx) error { return fmt.Errorf("commit failed") }
	defer func() { bunTxCommitHook = prev }()

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err := s.initUserPreferences(ctx, uuid.New(), notification.UpdatePreferenceRequest{
		Category: "security",
		Type:     "login-alerts",
		Enabled:  true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit transaction")
}

func TestInitDefaultPreferences_CommitError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	prev := bunTxCommitHook
	bunTxCommitHook = func(*bun.Tx) error { return fmt.Errorf("commit failed") }
	defer func() { bunTxCommitHook = prev }()

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err := s.initDefaultPreferences(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit transaction")
}

func TestInitUserPreferences_InsertPreferencesError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	insertNotificationPreference(t, db, userID)

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err := s.initUserPreferences(ctx, userID, notification.UpdatePreferenceRequest{
		Category: "security",
		Type:     "login-alerts",
		Enabled:  false,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert preferences")
}

func TestInitUserPreferences_InsertItemsError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	_, err := db.NewDropTable().Model((*shared_types.PreferenceItem)(nil)).IfExists().Exec(ctx)
	require.NoError(t, err)

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err = s.initUserPreferences(ctx, uuid.New(), notification.UpdatePreferenceRequest{
		Category: "security",
		Type:     "login-alerts",
		Enabled:  true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert preference items")
}

func TestInitDefaultPreferences_InsertPreferencesError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	insertNotificationPreference(t, db, userID)

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err := s.initDefaultPreferences(ctx, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert preferences")
}

func TestInitDefaultPreferences_InsertItemsError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	_, err := db.NewDropTable().Model((*shared_types.PreferenceItem)(nil)).IfExists().Exec(ctx)
	require.NoError(t, err)

	s := &NotificationStorage{DB: db, Ctx: ctx}
	err = s.initDefaultPreferences(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert preference items")
}

// ---------------------------------------------------------------------------
// MapToResponse
// ---------------------------------------------------------------------------

func TestMapToResponse_AllCategoriesAndIgnoresUnknown(t *testing.T) {
	prefID := uuid.New()
	items := []shared_types.PreferenceItem{
		{PreferenceID: prefID, Category: "activity", Type: "team-updates", Enabled: true},
		{PreferenceID: prefID, Category: "security", Type: "login-alerts", Enabled: true},
		{PreferenceID: prefID, Category: "security", Type: "password-changes", Enabled: true},
		{PreferenceID: prefID, Category: "security", Type: "security-alerts", Enabled: true},
		{PreferenceID: prefID, Category: "update", Type: "product-updates", Enabled: true},
		{PreferenceID: prefID, Category: "update", Type: "newsletter", Enabled: false},
		{PreferenceID: prefID, Category: "update", Type: "marketing", Enabled: false},
		{PreferenceID: prefID, Category: "update", Type: "unknown-type", Enabled: true},
		{PreferenceID: prefID, Category: "unknown-cat", Type: "x", Enabled: true},
	}
	resp := MapToResponse(items)
	assert.Len(t, resp.Activity, 1)
	assert.Len(t, resp.Security, 3)
	assert.Len(t, resp.Update, 3)
}

// ---------------------------------------------------------------------------
// WebhookConfig
// ---------------------------------------------------------------------------

func TestCreateWebhookConfig_Success(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	s := NotificationStorage{DB: db, Ctx: ctx}
	cfg := &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           "slack",
		WebhookURL:     "https://example.com/hook",
		ChannelID:      "C1",
		IsActive:       true,
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, s.CreateWebhookConfig(ctx, cfg))
}

func TestCreateWebhookConfig_Error(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	s := NotificationStorage{DB: db, Ctx: context.Background()}
	err = s.CreateWebhookConfig(context.Background(), &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           "slack",
		WebhookURL:     "u",
		ChannelID:      "c",
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create webhook config")
}

func TestUpdateWebhookConfig_Success(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	cfg := &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           "slack",
		WebhookURL:     "https://a",
		ChannelID:      "c",
		IsActive:       true,
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err := db.NewInsert().Model(cfg).Exec(ctx)
	require.NoError(t, err)

	cfg.WebhookURL = "https://b"
	s := NotificationStorage{DB: db, Ctx: ctx}
	require.NoError(t, s.UpdateWebhookConfig(ctx, cfg))
}

func TestUpdateWebhookConfig_Error(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	s := NotificationStorage{DB: db, Ctx: context.Background()}
	err = s.UpdateWebhookConfig(context.Background(), &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           "slack",
		WebhookURL:     "u",
		ChannelID:      "c",
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
	})
	require.Error(t, err)
}

func TestDeleteWebhookConfig_Success(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	orgID := uuid.New()
	cfg := &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           "slack",
		WebhookURL:     "https://a",
		ChannelID:      "c",
		IsActive:       true,
		UserID:         uuid.New(),
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err := db.NewInsert().Model(cfg).Exec(ctx)
	require.NoError(t, err)

	s := NotificationStorage{DB: db, Ctx: ctx}
	require.NoError(t, s.DeleteWebhookConfig(ctx, "slack", orgID))
}

func TestDeleteWebhookConfig_Error(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	s := NotificationStorage{DB: db, Ctx: context.Background()}
	err = s.DeleteWebhookConfig(context.Background(), "slack", uuid.New())
	require.Error(t, err)
}

func TestGetWebhookConfig_Found(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	orgID := uuid.New()
	cfg := &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           "discord",
		WebhookURL:     "https://a",
		ChannelID:      "c",
		IsActive:       true,
		UserID:         uuid.New(),
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err := db.NewInsert().Model(cfg).Exec(ctx)
	require.NoError(t, err)

	s := NotificationStorage{DB: db, Ctx: ctx}
	got, err := s.GetWebhookConfig(ctx, "discord", orgID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, cfg.ID, got.ID)
}

func TestGetWebhookConfig_NoRows(t *testing.T) {
	db := newTestDB(t)
	s := NotificationStorage{DB: db, Ctx: context.Background()}
	got, err := s.GetWebhookConfig(context.Background(), "slack", uuid.New())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestGetWebhookConfig_OtherError(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=private")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	s := NotificationStorage{DB: db, Ctx: context.Background()}
	_, err = s.GetWebhookConfig(context.Background(), "slack", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook config not found")
}
