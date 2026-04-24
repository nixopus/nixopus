package utils

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/auth"
	"github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func sessionWithOrg(id string) *auth.SessionResponse {
	var r auth.SessionResponse
	r.Session.ActiveOrganizationID = &id
	return &r
}

func testBunSQLite(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file:memorg?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	ctx := context.Background()
	for _, stmt := range []string{
		`CREATE TABLE organization (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			logo TEXT,
			created_at DATETIME NOT NULL,
			metadata TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS organization_slug_uq ON organization (slug)`,
		`CREATE TABLE organization_settings (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL UNIQUE,
			settings TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
	} {
		_, err = db.ExecContext(ctx, stmt)
		require.NoError(t, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func testApp(t *testing.T, db *bun.DB) *storage.App {
	t.Helper()
	return &storage.App{Store: &storage.Store{DB: db}}
}

func TestGetOrganizationIDFromBetterAuth(t *testing.T) {
	old := verifySessionFn
	t.Cleanup(func() { verifySessionFn = old })

	orgID := uuid.New().String()
	t.Run("active org in session", func(t *testing.T) {
		verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
			return sessionWithOrg(orgID), nil
		}
		got, err := GetOrganizationIDFromBetterAuth(httptestNewRequest(t))
		require.NoError(t, err)
		assert.Equal(t, orgID, got)
	})
	t.Run("fallback header", func(t *testing.T) {
		verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
			return &auth.SessionResponse{}, nil
		}
		r := httptestNewRequest(t)
		r.Header.Set("X-Organization-Id", "header-org")
		got, err := GetOrganizationIDFromBetterAuth(r)
		require.NoError(t, err)
		assert.Equal(t, "header-org", got)
	})
	t.Run("verify error", func(t *testing.T) {
		verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
			return nil, errors.New("nope")
		}
		_, err := GetOrganizationIDFromBetterAuth(httptestNewRequest(t))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify session")
	})
}

func httptestNewRequest(t *testing.T) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	return r
}

func TestGetOrCreateOrganizationID_fromContext(t *testing.T) {
	id := uuid.New()
	t.Run("string in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, id.String())
		got, err := GetOrCreateOrganizationID(ctx, httptestNewRequest(t), nil)
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})
	t.Run("uuid in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, id)
		got, err := GetOrCreateOrganizationID(ctx, httptestNewRequest(t), nil)
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})
	// Unparseable string in context should fall through to Better Auth, not return early.
	t.Run("invalid string in context uses session", func(t *testing.T) {
		old := verifySessionFn
		t.Cleanup(func() { verifySessionFn = old })
		db := testBunSQLite(t)
		app := testApp(t, db)
		sessID := uuid.New()
		verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
			return sessionWithOrg(sessID.String()), nil
		}
		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "not-a-valid-uuid")
		got, err := GetOrCreateOrganizationID(ctx, httptestNewRequest(t), app)
		require.NoError(t, err)
		assert.Equal(t, sessID, got)
	})
}

func TestGetOrCreateOrganizationID_sessionAndDB(t *testing.T) {
	old := verifySessionFn
	t.Cleanup(func() { verifySessionFn = old })

	db := testBunSQLite(t)
	app := testApp(t, db)
	oid := uuid.New()
	oidStr := oid.String()

	verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
		return sessionWithOrg(oidStr), nil
	}

	ctx := context.Background()
	got, err := GetOrCreateOrganizationID(ctx, httptestNewRequest(t), app)
	require.NoError(t, err)
	assert.Equal(t, oid, got)

	var count int
	err = db.NewRaw("SELECT COUNT(*) FROM organization WHERE id = ?", oidStr).Scan(ctx, &count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestGetOrCreateOrganizationID_idempotentWhenOrgExists(t *testing.T) {
	old := verifySessionFn
	t.Cleanup(func() { verifySessionFn = old })

	db := testBunSQLite(t)
	app := testApp(t, db)
	oid := uuid.New()
	oidStr := oid.String()

	existing := types.Organization{
		ID:        oid,
		Name:      "Existing",
		Slug:      "existing",
		CreatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(&existing).Exec(context.Background())
	require.NoError(t, err)

	verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
		return sessionWithOrg(oidStr), nil
	}
	_, err = GetOrCreateOrganizationID(context.Background(), httptestNewRequest(t), app)
	require.NoError(t, err)
}

func TestGetOrCreateOrganizationID_notFound(t *testing.T) {
	old := verifySessionFn
	t.Cleanup(func() { verifySessionFn = old })
	verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
		return &auth.SessionResponse{}, nil
	}
	_, err := GetOrCreateOrganizationID(context.Background(), httptestNewRequest(t), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization ID not found")
}

func TestGetOrCreateOrganizationID_invalidUUIDFromSession(t *testing.T) {
	old := verifySessionFn
	t.Cleanup(func() { verifySessionFn = old })
	bad := "not-a-uuid"
	verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
		return sessionWithOrg(bad), nil
	}
	_, err := GetOrCreateOrganizationID(context.Background(), httptestNewRequest(t), nil)
	require.Error(t, err)
}

// Two org UUIDs whose string forms share the first 8 characters (the slug) collide on the unique index.
var slugCollisionID1 = uuid.MustParse("11111111-0000-4000-8000-000000000001")
var slugCollisionID2 = uuid.MustParse("11111111-0000-4000-8000-000000000002")

func TestEnsureOrganizationExists_invalidUUIDString(t *testing.T) {
	db := testBunSQLite(t)
	app := testApp(t, db)
	err := ensureOrganizationExists(context.Background(), app, "totally-not-a-uuid")
	require.Error(t, err)
}

func TestEnsureOrganizationExists_uniqueSlugFromEnsureFails(t *testing.T) {
	db := testBunSQLite(t)
	app := testApp(t, db)
	ctx := context.Background()

	first := types.Organization{
		ID:        slugCollisionID1,
		Name:      "A",
		Slug:      slugCollisionID1.String()[:8],
		CreatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(&first).Exec(ctx)
	require.NoError(t, err)

	err = ensureOrganizationExists(ctx, app, slugCollisionID2.String())
	require.Error(t, err)
}

func TestGetOrCreateOrganizationID_ensureFails(t *testing.T) {
	old := verifySessionFn
	t.Cleanup(func() { verifySessionFn = old })

	db := testBunSQLite(t)
	app := testApp(t, db)
	ctx := context.Background()

	existing := types.Organization{
		ID:        slugCollisionID1,
		Name:      "A",
		Slug:      slugCollisionID1.String()[:8],
		CreatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(&existing).Exec(ctx)
	require.NoError(t, err)

	verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
		return sessionWithOrg(slugCollisionID2.String()), nil
	}
	_, err = GetOrCreateOrganizationID(ctx, httptestNewRequest(t), app)
	require.Error(t, err)
}

func TestGetOrganizationIDString(t *testing.T) {
	old := verifySessionFn
	t.Cleanup(func() { verifySessionFn = old })

	id := uuid.New()
	t.Run("from context string", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, id.String())
		s, err := GetOrganizationIDString(ctx, httptestNewRequest(t), nil)
		require.NoError(t, err)
		assert.Equal(t, id.String(), s)
	})
	t.Run("from context uuid", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, id)
		s, err := GetOrganizationIDString(ctx, httptestNewRequest(t), nil)
		require.NoError(t, err)
		assert.Equal(t, id.String(), s)
	})
	t.Run("from session with app ensure", func(t *testing.T) {
		db := testBunSQLite(t)
		app := testApp(t, db)
		oidStr := id.String()
		verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
			return sessionWithOrg(oidStr), nil
		}
		s, err := GetOrganizationIDString(context.Background(), httptestNewRequest(t), app)
		require.NoError(t, err)
		assert.Equal(t, oidStr, s)
	})
	t.Run("verify fails", func(t *testing.T) {
		verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
			return nil, errors.New("x")
		}
		_, err := GetOrganizationIDString(context.Background(), httptestNewRequest(t), nil)
		require.Error(t, err)
	})
	t.Run("empty string in context falls through to session", func(t *testing.T) {
		verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
			return sessionWithOrg(id.String()), nil
		}
		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "")
		s, err := GetOrganizationIDString(ctx, httptestNewRequest(t), nil)
		require.NoError(t, err)
		assert.Equal(t, id.String(), s)
	})
	t.Run("ensure error when app is set", func(t *testing.T) {
		db := testBunSQLite(t)
		app := testApp(t, db)
		ctx := context.Background()
		_, err := db.NewInsert().Model(&types.Organization{
			ID:        slugCollisionID1,
			Name:      "A",
			Slug:      slugCollisionID1.String()[:8],
			CreatedAt: time.Now(),
		}).Exec(ctx)
		require.NoError(t, err)

		verifySessionFn = func(r *http.Request) (*auth.SessionResponse, error) {
			return sessionWithOrg(slugCollisionID2.String()), nil
		}
		_, err = GetOrganizationIDString(context.Background(), httptestNewRequest(t), app)
		require.Error(t, err)
	})
}

func TestGetOrganizationSettings_createsWhenMissing(t *testing.T) {
	db := testBunSQLite(t)
	ctx := context.Background()
	orgID := uuid.New()
	data, err := GetOrganizationSettings(ctx, db, orgID)
	require.NoError(t, err)
	assert.NotNil(t, data.ContainerLogTailLines)
	var n int
	err = db.NewRaw("SELECT COUNT(*) FROM organization_settings WHERE organization_id = ?", orgID.String()).Scan(ctx, &n)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestGetOrganizationSettings_readsExisting(t *testing.T) {
	db := testBunSQLite(t)
	ctx := context.Background()
	oid := uuid.New()
	lines := 42
	settings := types.OrganizationSettings{
		ID:             uuid.New(),
		OrganizationID: oid,
		Settings: types.OrganizationSettingsData{
			ContainerLogTailLines: &lines,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(&settings).Exec(ctx)
	require.NoError(t, err)
	data, err := GetOrganizationSettings(ctx, db, oid)
	require.NoError(t, err)
	require.NotNil(t, data.ContainerLogTailLines)
	assert.Equal(t, 42, *data.ContainerLogTailLines)
}

func TestGetOrganizationSettings_nonNoRowsErrorReturnsDefaults(t *testing.T) {
	db := testBunSQLite(t)
	_ = db.Close()
	data, err := GetOrganizationSettings(context.Background(), db, uuid.New())
	require.NoError(t, err)
	_ = data
}

func TestGetOrganizationSettings_insertLockedThenRereadEmptyReturnsDefaults(t *testing.T) {
	db := testBunSQLite(t)
	ctx := context.Background()
	orgID := uuid.New()
	_, err := db.ExecContext(ctx, `CREATE TRIGGER tr_reject_osa_settings_before_insert BEFORE INSERT ON organization_settings
		BEGIN SELECT RAISE(ABORT, 'rejected insert'); END;`)
	require.NoError(t, err)

	data, err := GetOrganizationSettings(ctx, db, orgID)
	require.NoError(t, err)
	_ = data
}

func TestGetOrganizationSettings_insertLocked_hookThenMerges(t *testing.T) {
	db := testBunSQLite(t)
	ctx := context.Background()
	orgID := uuid.New()
	_, err := db.ExecContext(ctx, `CREATE TRIGGER tr_reject_osa_settings_bi BEFORE INSERT ON organization_settings
		BEGIN SELECT RAISE(ABORT, 'rejected insert'); END;`)
	require.NoError(t, err)

	var hookRan bool
	old := getOrganizationSettingsAfterFailedInsert
	t.Cleanup(func() { getOrganizationSettingsAfterFailedInsert = old })
	getOrganizationSettingsAfterFailedInsert = func(c context.Context, b *bun.DB, id uuid.UUID) {
		hookRan = true
		_, e := b.ExecContext(c, `DROP TRIGGER tr_reject_osa_settings_bi`)
		require.NoError(t, e)
		row := types.OrganizationSettings{
			ID:             uuid.New(),
			OrganizationID: id,
			Settings:       types.DefaultOrganizationSettingsData(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		_, e = b.NewInsert().Model(&row).Exec(c)
		require.NoError(t, e)
	}

	data, err := GetOrganizationSettings(ctx, db, orgID)
	require.NoError(t, err)
	assert.True(t, hookRan)
	assert.NotNil(t, data.ContainerLogTailLines)
}

func TestGetOrganizationSettings_mergesAllOptionalPointers(t *testing.T) {
	db := testBunSQLite(t)
	ctx := context.Background()
	oid := uuid.New()
	deployEn := true
	deployRet := 14
	auditEn := true
	auditRet := 21
	extEn := true
	extRet := 9
	backupOn := true
	freq := "weekly"
	backupHour := 3
	backupDow := 2
	backupN := 4
	ai := true
	s3 := true
	lines := 200
	policy := "no"
	stopt := 30
	aprI := true
	aprB := true
	settings := types.OrganizationSettings{
		ID:             uuid.New(),
		OrganizationID: oid,
		Settings: types.OrganizationSettingsData{
			WebsocketReconnectAttempts:       7,
			WebsocketReconnectInterval:       6,
			ApiRetryAttempts:                 2,
			DisableApiCache:                  true,
			ContainerLogTailLines:            &lines,
			ContainerDefaultRestartPolicy:    &policy,
			ContainerStopTimeout:             &stopt,
			ContainerAutoPruneDanglingImages: &aprI,
			ContainerAutoPruneBuildCache:     &aprB,
			DeploymentLogsCleanupEnabled:     &deployEn,
			DeploymentLogsRetentionDays:      &deployRet,
			AuditLogsCleanupEnabled:          &auditEn,
			AuditLogsRetentionDays:           &auditRet,
			ExtensionLogsCleanupEnabled:      &extEn,
			ExtensionLogsRetentionDays:       &extRet,
			BackupScheduleEnabled:            &backupOn,
			BackupScheduleFrequency:          &freq,
			BackupScheduleHourUTC:            &backupHour,
			BackupScheduleDayOfWeek:          &backupDow,
			BackupRetentionCount:             &backupN,
			AIIncidentsEnabled:               &ai,
			S3ArtifactUploadEnabled:          &s3,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(&settings).Exec(ctx)
	require.NoError(t, err)
	data, err := GetOrganizationSettings(ctx, db, oid)
	require.NoError(t, err)
	require.NotNil(t, data.ContainerLogTailLines)
	assert.Equal(t, 200, *data.ContainerLogTailLines)
	assert.Equal(t, "no", *data.ContainerDefaultRestartPolicy)
	assert.Equal(t, 30, *data.ContainerStopTimeout)
	assert.True(t, *data.ContainerAutoPruneDanglingImages)
	assert.True(t, *data.ContainerAutoPruneBuildCache)
	assert.Equal(t, 7, data.WebsocketReconnectAttempts)
	assert.Equal(t, 6, data.WebsocketReconnectInterval)
	assert.Equal(t, 2, data.ApiRetryAttempts)
	assert.True(t, data.DisableApiCache)
	require.NotNil(t, data.DeploymentLogsCleanupEnabled)
	assert.True(t, *data.DeploymentLogsCleanupEnabled)
	assert.Equal(t, 14, *data.DeploymentLogsRetentionDays)
	require.NotNil(t, data.AuditLogsCleanupEnabled)
	assert.True(t, *data.AuditLogsCleanupEnabled)
	assert.Equal(t, 21, *data.AuditLogsRetentionDays)
	require.NotNil(t, data.ExtensionLogsCleanupEnabled)
	assert.True(t, *data.ExtensionLogsCleanupEnabled)
	assert.Equal(t, 9, *data.ExtensionLogsRetentionDays)
	require.NotNil(t, data.BackupScheduleEnabled)
	assert.True(t, *data.BackupScheduleEnabled)
	assert.Equal(t, "weekly", *data.BackupScheduleFrequency)
	assert.Equal(t, 3, *data.BackupScheduleHourUTC)
	assert.Equal(t, 2, *data.BackupScheduleDayOfWeek)
	assert.Equal(t, 4, *data.BackupRetentionCount)
	require.NotNil(t, data.AIIncidentsEnabled)
	assert.True(t, *data.AIIncidentsEnabled)
	require.NotNil(t, data.S3ArtifactUploadEnabled)
	assert.True(t, *data.S3ArtifactUploadEnabled)
}
