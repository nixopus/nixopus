package storage_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/storage"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func sqliteTelemetryDB(t *testing.T) (*bun.DB, context.Context) {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file:memtelemetry"+uuid.New().String()+"?mode=memory&cache=shared")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	ctx := context.Background()

	db.RegisterModel((*types.CliInstallation)(nil))

	_, err = db.ExecContext(ctx, `CREATE TABLE cli_installations (
		id         TEXT    PRIMARY KEY,
		event_type TEXT    NOT NULL,
		os         TEXT    NOT NULL DEFAULT 'unknown',
		arch       TEXT    NOT NULL DEFAULT 'unknown',
		version    TEXT    NOT NULL,
		duration   INTEGER NOT NULL DEFAULT 0,
		error      TEXT,
		ip_hash    TEXT    NOT NULL,
		created_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	t.Cleanup(func() { _ = db.Close() })
	return db, ctx
}

func TestNewTelemetryStorage(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteTelemetryDB(t)
	s := storage.NewTelemetryStorage(db, ctx, nil)
	require.NotNil(t, s)
	assert.Equal(t, db, s.DB)
	assert.Equal(t, ctx, s.Ctx)
}

func TestCreateInstallEvent_success(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteTelemetryDB(t)
	s := storage.NewTelemetryStorage(db, ctx, nil)

	event := &types.CliInstallation{
		ID:        uuid.New(),
		EventType: "install_success",
		OS:        "ubuntu",
		Arch:      "amd64",
		Version:   "1.0.0",
		Duration:  30,
		IPHash:    "abc123deadbeef",
		CreatedAt: time.Now(),
	}

	err := s.CreateInstallEvent(event)
	require.NoError(t, err)

	var count int
	err = db.NewRaw("SELECT COUNT(*) FROM cli_installations WHERE id = ?", event.ID.String()).Scan(ctx, &count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCreateInstallEvent_dbError(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteTelemetryDB(t)
	require.NoError(t, db.Close())
	s := storage.NewTelemetryStorage(db, ctx, nil)

	err := s.CreateInstallEvent(&types.CliInstallation{ID: uuid.New()})
	require.Error(t, err)
}
