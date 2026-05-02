package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func sqliteAuditDB(t *testing.T) (*bun.DB, context.Context) {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file:memaudit"+uuid.New().String()+"?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	ctx := context.Background()

	db.RegisterModel((*types.Organization)(nil))
	db.RegisterModel((*types.Member)(nil))
	db.RegisterModel((*types.OrganizationUsers)(nil))
	db.RegisterModel((*types.User)(nil))
	db.RegisterModel((*types.AuditLog)(nil))

	for _, stmt := range []string{
		`CREATE TABLE "user" (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			email_verified INTEGER NOT NULL DEFAULT 0,
			image TEXT,
			is_onboarded INTEGER NOT NULL DEFAULT 0,
			provision_status TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE organization (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			logo TEXT,
			created_at TEXT NOT NULL,
			metadata TEXT
		)`,
		`CREATE TABLE member (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE audit_logs (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			organization_id TEXT NOT NULL,
			action TEXT NOT NULL,
			resource_type TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			old_values TEXT,
			new_values TEXT,
			metadata TEXT,
			ip_address TEXT,
			user_agent TEXT,
			created_at TEXT NOT NULL,
			request_id TEXT
		)`,
	} {
		_, err := db.ExecContext(ctx, stmt)
		require.NoError(t, err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db, ctx
}

func TestNewAuditStorage(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuditDB(t)
	s := NewAuditStorage(db, ctx)
	require.NotNil(t, s)
	assert.Equal(t, db, s.DB)
	assert.Equal(t, ctx, s.Ctx)
}

func TestAuditStorage_roundTrip(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuditDB(t)
	s := NewAuditStorage(db, ctx)

	userID := uuid.New()
	orgID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.ExecContext(ctx,
		`INSERT INTO "user" (id, name, email, email_verified, is_onboarded, created_at, updated_at) VALUES (?, ?, ?, 0, 0, ?, ?)`,
		userID.String(), "name", "u@test", now, now,
	)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO organization (id, name, slug, created_at) VALUES (?, ?, ?, ?)`,
		orgID.String(), "Org", "slug", now,
	)
	require.NoError(t, err)

	log := &types.AuditLog{
		ID:             uuid.New(),
		UserID:         userID,
		OrganizationID: orgID,
		Action:         types.AuditActionCreate,
		ResourceType:   types.AuditResourceUser,
		ResourceID:     uuid.New(),
		OldValues:      map[string]any{"a": 1},
		NewValues:      map[string]any{"b": 2},
		IPAddress:      "127.0.0.1",
		UserAgent:      "ua",
		CreatedAt:      time.Now(),
		RequestID:      uuid.New(),
	}
	require.NoError(t, s.CreateAuditLog(log))

	got, total, err := s.GetAuditLogs(map[string]any{"organization_id": orgID}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, got, 1)
	assert.NotNil(t, got[0].User)
	assert.Equal(t, "u@test", got[0].User.Email)
	assert.NotNil(t, got[0].Organization)

	_, _, err = s.GetAuditLogs(map[string]any{"user_id": userID, "organization_id": orgID}, 1, 2)
	require.NoError(t, err)
}

func TestAuditStorage_GetAuditLogs_scanError(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuditDB(t)
	s := NewAuditStorage(db, ctx)

	userID := uuid.New()
	orgID := uuid.New()
	rid := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.ExecContext(ctx,
		`INSERT INTO "user" (id, name, email, email_verified, is_onboarded, created_at, updated_at) VALUES (?, ?, ?, 0, 0, ?, ?)`,
		userID.String(), "name", "badscan@test", now, now,
	)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO organization (id, name, slug, created_at) VALUES (?, ?, ?, ?)`,
		orgID.String(), "Org", "slug2", now,
	)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO audit_logs (id, user_id, organization_id, action, resource_type, resource_id, old_values, new_values, metadata, created_at, request_id)
		 VALUES (?, ?, ?, 'create', 'user', ?, 'not-json', '{}', '{}', ?, ?)`,
		uuid.New().String(), userID.String(), orgID.String(), rid.String(), now, uuid.New().String(),
	)
	require.NoError(t, err)

	_, _, err = s.GetAuditLogs(map[string]any{"organization_id": orgID}, 1, 10)
	require.Error(t, err)
}

func TestAuditStorage_CreateAuditLog_error(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuditDB(t)
	require.NoError(t, db.Close())
	s := NewAuditStorage(db, ctx)
	err := s.CreateAuditLog(&types.AuditLog{})
	require.Error(t, err)
}

func TestAuditStorage_GetAuditLogs_queryError(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuditDB(t)
	require.NoError(t, db.Close())
	s := NewAuditStorage(db, ctx)
	_, _, err := s.GetAuditLogs(nil, 1, 10)
	require.Error(t, err)
}
