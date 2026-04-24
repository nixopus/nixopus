package queue

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	apitypes "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func testBackupSQLite(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file:membackup?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	ctx := context.Background()
	_, err = db.ExecContext(ctx, `CREATE TABLE machine_backups (
		id TEXT PRIMARY KEY,
		status TEXT NOT NULL,
		updated_at DATETIME,
		started_at DATETIME,
		completed_at DATETIME,
		snapshot_path TEXT,
		s3_path TEXT,
		size_bytes INTEGER,
		error TEXT
	)`)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func withBackupTestConfig(t *testing.T) func() {
	t.Helper()
	orig := config.AppConfig
	config.AppConfig.BetterAuth.Secret = "unit-test-secret-key-32bytes!xx"
	config.AppConfig.S3 = apitypes.S3Config{
		Endpoint:  "localhost:9000",
		Bucket:    "bucket",
		Region:    "",
		AccessKey: "ak",
		SecretKey: "sk",
		UseSSL:    false,
	}
	return func() { config.AppConfig = orig }
}

type mockBackupSSH struct {
	fn func(cmd string) (string, error)
}

func (m *mockBackupSSH) RunCommand(cmd string) (string, error) {
	if m.fn != nil {
		return m.fn(cmd)
	}
	return "", nil
}

func TestBackupWorker_Handle(t *testing.T) {
	ctx := context.Background()
	prev := getSSHManagerForServerForBackup
	t.Cleanup(func() { getSSHManagerForServerForBackup = prev })

	t.Run("expired", func(t *testing.T) {
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		err := w.Handle(ctx, MachineBackupPayload{
			OrgID:       uuid.New().String(),
			ServerID:    uuid.New().String(),
			BackupRowID: uuid.New().String(),
			ExpiresAt:   time.Now().Add(-time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("invalid backup row id", func(t *testing.T) {
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		err := w.Handle(ctx, MachineBackupPayload{
			OrgID:       uuid.New().String(),
			ServerID:    uuid.New().String(),
			BackupRowID: "bad",
			ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("invalid org id", func(t *testing.T) {
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		err := w.Handle(ctx, MachineBackupPayload{
			OrgID:       "bad-org",
			ServerID:    uuid.New().String(),
			BackupRowID: uuid.New().String(),
			ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("invalid server id", func(t *testing.T) {
		done := withBackupTestConfig(t)
		defer done()
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		err = w.Handle(ctx, MachineBackupPayload{
			OrgID:       uuid.New().String(),
			ServerID:    "bad-server",
			BackupRowID: row.String(),
			ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("in progress update fails", func(t *testing.T) {
		done := withBackupTestConfig(t)
		defer done()
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		_, _ = db.ExecContext(ctx, `DROP TABLE machine_backups`)
		getSSHManagerForServerForBackup = func(context.Context, uuid.UUID, uuid.UUID) (backupSSHRunner, error) {
			return &mockBackupSSH{}, nil
		}
		err = w.Handle(ctx, MachineBackupPayload{
			OrgID:       uuid.New().String(),
			ServerID:    uuid.New().String(),
			BackupRowID: row.String(),
			ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("ssh manager error", func(t *testing.T) {
		done := withBackupTestConfig(t)
		defer done()
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		getSSHManagerForServerForBackup = func(context.Context, uuid.UUID, uuid.UUID) (backupSSHRunner, error) {
			return nil, errors.New("no ssh")
		}
		err = w.Handle(ctx, MachineBackupPayload{
			OrgID:       uuid.New().String(),
			ServerID:    uuid.New().String(),
			BackupRowID: row.String(),
			ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("install restic fails", func(t *testing.T) {
		done := withBackupTestConfig(t)
		defer done()
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		getSSHManagerForServerForBackup = func(context.Context, uuid.UUID, uuid.UUID) (backupSSHRunner, error) {
			return &mockBackupSSH{fn: func(cmd string) (string, error) {
				if strings.Contains(cmd, "which restic >/dev/null") {
					return "", errors.New("install failed")
				}
				return "", nil
			}}, nil
		}
		err = w.Handle(ctx, MachineBackupPayload{
			OrgID:       uuid.New().String(),
			ServerID:    uuid.New().String(),
			BackupRowID: row.String(),
			ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("restic init fails", func(t *testing.T) {
		done := withBackupTestConfig(t)
		defer done()
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		getSSHManagerForServerForBackup = func(context.Context, uuid.UUID, uuid.UUID) (backupSSHRunner, error) {
			return &mockBackupSSH{fn: func(cmd string) (string, error) {
				if strings.Contains(cmd, "which restic >/dev/null") {
					return "", nil
				}
				if strings.Contains(cmd, "curl -fsSL") {
					return "", nil
				}
				if strings.Contains(cmd, "which restic 2>/dev/null") {
					return "/usr/bin/restic", nil
				}
				if strings.Contains(cmd, " init ") {
					return "", errors.New("init boom")
				}
				return "", nil
			}}, nil
		}
		err = w.Handle(ctx, MachineBackupPayload{
			OrgID:       uuid.New().String(),
			ServerID:    uuid.New().String(),
			BackupRowID: row.String(),
			ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("backup command fails with long output", func(t *testing.T) {
		done := withBackupTestConfig(t)
		defer done()
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		longErr := strings.Repeat("E", 600)
		getSSHManagerForServerForBackup = func(context.Context, uuid.UUID, uuid.UUID) (backupSSHRunner, error) {
			return &mockBackupSSH{fn: func(cmd string) (string, error) {
				if strings.Contains(cmd, "which restic >/dev/null") {
					return "", nil
				}
				if strings.Contains(cmd, "curl -fsSL") {
					return "", nil
				}
				if strings.Contains(cmd, "which restic 2>/dev/null") {
					return "/usr/bin/restic", nil
				}
				if strings.Contains(cmd, " init ") {
					return "", nil
				}
				if strings.Contains(cmd, " backup ") {
					return longErr, errors.New("backup failed")
				}
				return "", nil
			}}, nil
		}
		err = w.Handle(ctx, MachineBackupPayload{
			OrgID:       uuid.New().String(),
			ServerID:    uuid.New().String(),
			BackupRowID: row.String(),
			ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("prune warning non fatal", func(t *testing.T) {
		done := withBackupTestConfig(t)
		defer done()
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		summary := `{"message_type":"summary","snapshot_id":"snapw","total_bytes_processed":5}`
		getSSHManagerForServerForBackup = func(context.Context, uuid.UUID, uuid.UUID) (backupSSHRunner, error) {
			return &mockBackupSSH{fn: func(cmd string) (string, error) {
				if strings.Contains(cmd, "which restic >/dev/null") {
					return "", nil
				}
				if strings.Contains(cmd, "curl -fsSL") {
					return "", nil
				}
				if strings.Contains(cmd, "which restic 2>/dev/null") {
					return "/usr/bin/restic", nil
				}
				if strings.Contains(cmd, " init ") {
					return "", nil
				}
				if strings.Contains(cmd, " backup ") {
					return summary, nil
				}
				if strings.Contains(cmd, " forget ") {
					return "", errors.New("prune warn")
				}
				return "", nil
			}}, nil
		}
		err = w.Handle(ctx, MachineBackupPayload{
			OrgID:          uuid.New().String(),
			ServerID:       uuid.New().String(),
			BackupRowID:    row.String(),
			RetentionCount: 0,
			ExpiresAt:      time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("completed_update_fails_check_constraint", func(t *testing.T) {
		done := withBackupTestConfig(t)
		defer done()
		sqldb, err := sql.Open("sqlite", "file:memchk?mode=memory&cache=shared&_pragma=foreign_keys(1)")
		require.NoError(t, err)
		db := bun.NewDB(sqldb, sqlitedialect.New())
		_, err = db.ExecContext(ctx, `CREATE TABLE machine_backups (
			id TEXT PRIMARY KEY,
			status TEXT NOT NULL CHECK (status IN ('pending','in_progress','failed')),
			updated_at DATETIME,
			started_at DATETIME,
			completed_at DATETIME,
			snapshot_path TEXT,
			s3_path TEXT,
			size_bytes INTEGER,
			error TEXT
		)`)
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err = db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		summary := `{"message_type":"summary","snapshot_id":"snapchk","total_bytes_processed":1}`
		getSSHManagerForServerForBackup = func(context.Context, uuid.UUID, uuid.UUID) (backupSSHRunner, error) {
			return &mockBackupSSH{fn: func(cmd string) (string, error) {
				if strings.Contains(cmd, "which restic >/dev/null") {
					return "", nil
				}
				if strings.Contains(cmd, "curl -fsSL") {
					return "", nil
				}
				if strings.Contains(cmd, "which restic 2>/dev/null") {
					return "/usr/bin/restic", nil
				}
				if strings.Contains(cmd, " init ") {
					return "", nil
				}
				if strings.Contains(cmd, " backup ") {
					return summary, nil
				}
				if strings.Contains(cmd, " forget ") {
					return "", nil
				}
				return "", nil
			}}, nil
		}
		err = w.Handle(ctx, MachineBackupPayload{
			OrgID:       uuid.New().String(),
			ServerID:    uuid.New().String(),
			BackupRowID: row.String(),
			ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
	})

	t.Run("full success", func(t *testing.T) {
		done := withBackupTestConfig(t)
		defer done()
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		summary := `{"message_type":"summary","snapshot_id":"snapok","total_bytes_processed":42}`
		getSSHManagerForServerForBackup = func(context.Context, uuid.UUID, uuid.UUID) (backupSSHRunner, error) {
			return &mockBackupSSH{fn: func(cmd string) (string, error) {
				if strings.Contains(cmd, "which restic >/dev/null") {
					return "", nil
				}
				if strings.Contains(cmd, "curl -fsSL") {
					return "", nil
				}
				if strings.Contains(cmd, "which restic 2>/dev/null") {
					return "/usr/bin/restic", nil
				}
				if strings.Contains(cmd, " init ") {
					return "", nil
				}
				if strings.Contains(cmd, " backup ") {
					return summary, nil
				}
				if strings.Contains(cmd, " forget ") {
					return "", nil
				}
				return "", nil
			}}, nil
		}
		err = w.Handle(ctx, MachineBackupPayload{
			OrgID:          uuid.New().String(),
			ServerID:       uuid.New().String(),
			BackupRowID:    row.String(),
			BackupPaths:    nil,
			RetentionCount: 3,
			ExpiresAt:      time.Now().Add(time.Hour).Unix(),
		})
		require.NoError(t, err)
		var status string
		require.NoError(t, db.QueryRowContext(ctx, `SELECT status FROM machine_backups WHERE id = ?`, row.String()).Scan(&status))
		require.Equal(t, string(machine_types.BackupStatusCompleted), status)
	})

	t.Run("failBackup", func(t *testing.T) {
		db := testBackupSQLite(t)
		w := &BackupWorker{db: db, backupStore: machine_storage.NewBackupStorage(db, ctx)}
		row := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO machine_backups (id, status, updated_at) VALUES (?, 'pending', datetime('now'))`, row.String())
		require.NoError(t, err)
		require.NoError(t, w.failBackup(ctx, row, "boom"))
		var st string
		var errMsg string
		require.NoError(t, db.QueryRowContext(ctx, `SELECT status, error FROM machine_backups WHERE id = ?`, row.String()).Scan(&st, &errMsg))
		require.Equal(t, string(machine_types.BackupStatusFailed), st)
		require.Equal(t, "boom", errMsg)
	})
}

func TestEnqueueMachineBackup_errors(t *testing.T) {
	ctx := context.Background()
	_, err := EnqueueMachineBackup(ctx, MachineBackupPayload{})
	require.Error(t, err)
}
