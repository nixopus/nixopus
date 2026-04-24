package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func testSQLiteDB(t *testing.T) (*bun.DB, context.Context) {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file:mwtest?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	db.RegisterModel((*types.Organization)(nil))
	db.RegisterModel((*types.Member)(nil))
	db.RegisterModel((*types.OrganizationUsers)(nil))
	db.RegisterModel((*types.User)(nil))
	ctx := context.Background()
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
		`CREATE TABLE feature_flags (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			feature_name TEXT NOT NULL,
			is_enabled INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			deleted_at TEXT
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
		_, err = db.ExecContext(ctx, stmt)
		require.NoError(t, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, ctx
}

func testApp(t *testing.T) *storage.App {
	t.Helper()
	db, ctx := testSQLiteDB(t)
	return &storage.App{Store: &storage.Store{DB: db}, Ctx: ctx}
}

func insertUser(t *testing.T, db *bun.DB, ctx context.Context, id uuid.UUID, email string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.ExecContext(ctx,
		`INSERT INTO "user" (id, name, email, email_verified, is_onboarded, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id.String(), "n", email, 1, 0, now, now,
	)
	require.NoError(t, err)
}

func insertMember(t *testing.T, db *bun.DB, ctx context.Context, userID, orgID uuid.UUID) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.ExecContext(ctx,
		`INSERT INTO member (id, organization_id, user_id, role, created_at) VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), orgID.String(), userID.String(), "member", now,
	)
	require.NoError(t, err)
}

func clearRateLimiterClients(t *testing.T) {
	t.Helper()
	clientsMtx.Lock()
	clients = make(map[string]*client)
	clientsMtx.Unlock()
}

func resetJWKSState(t *testing.T) {
	t.Helper()
	jwksMu.Lock()
	jwksCache = nil
	jwksExpiry = time.Time{}
	jwksMu.Unlock()
}

// roundTripFunc implements http.RoundTripper for tests.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
