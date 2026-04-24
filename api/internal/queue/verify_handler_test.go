package queue

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	cryptossh "golang.org/x/crypto/ssh"
	_ "modernc.org/sqlite"
)

func testVerifySQLite(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file:memverify?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	ctx := context.Background()
	for _, stmt := range []string{
		`CREATE TABLE ssh_keys (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			host TEXT,
			proxy_host TEXT,
			user TEXT,
			port INTEGER,
			public_key TEXT,
			private_key_encrypted TEXT,
			password_encrypted TEXT,
			key_type TEXT,
			key_size INTEGER,
			fingerprint TEXT,
			auth_method TEXT NOT NULL DEFAULT 'key',
			is_active INTEGER NOT NULL DEFAULT 1,
			is_default INTEGER NOT NULL DEFAULT 0,
			last_used_at DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME
		)`,
	} {
		_, err = db.ExecContext(ctx, stmt)
		require.NoError(t, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func rsaPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	blk := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return string(pem.EncodeToMemory(blk))
}

func Test_handleMachineVerify(t *testing.T) {
	prevProbe := machineVerifySSHProbe
	t.Cleanup(func() { machineVerifySSHProbe = prevProbe })

	ctx := context.Background()

	t.Run("invalid machine_id", func(t *testing.T) {
		db := testVerifySQLite(t)
		err := handleMachineVerify(ctx, db, MachineVerifyPayload{MachineID: "x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid machine_id")
	})

	t.Run("ssh key not found", func(t *testing.T) {
		db := testVerifySQLite(t)
		err := handleMachineVerify(ctx, db, MachineVerifyPayload{MachineID: uuid.New().String()})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load ssh key")
	})

	t.Run("missing private key or host", func(t *testing.T) {
		db := testVerifySQLite(t)
		id := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO ssh_keys (id, organization_id, name, host, private_key_encrypted, is_active, created_at, updated_at) VALUES (?, ?, 'n', NULL, NULL, 1, datetime('now'), datetime('now'))`,
			id.String(), uuid.New().String())
		require.NoError(t, err)
		err = handleMachineVerify(ctx, db, MachineVerifyPayload{MachineID: id.String()})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing private key or host")
	})

	t.Run("invalid private key pem", func(t *testing.T) {
		db := testVerifySQLite(t)
		id := uuid.New()
		host := "127.0.0.1"
		bad := "not-a-key"
		_, err := db.ExecContext(ctx, `INSERT INTO ssh_keys (id, organization_id, name, host, private_key_encrypted, is_active, created_at, updated_at) VALUES (?, ?, 'n', ?, ?, 1, datetime('now'), datetime('now'))`,
			id.String(), uuid.New().String(), host, bad)
		require.NoError(t, err)
		err = handleMachineVerify(ctx, db, MachineVerifyPayload{MachineID: id.String()})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})

	t.Run("ssh probe failure", func(t *testing.T) {
		db := testVerifySQLite(t)
		id := uuid.New()
		host := "127.0.0.1"
		pemData := rsaPEM(t)
		_, err := db.ExecContext(ctx, `INSERT INTO ssh_keys (id, organization_id, name, host, private_key_encrypted, is_active, created_at, updated_at) VALUES (?, ?, 'n', ?, ?, 1, datetime('now'), datetime('now'))`,
			id.String(), uuid.New().String(), host, pemData)
		require.NoError(t, err)
		machineVerifySSHProbe = func(context.Context, string, *cryptossh.ClientConfig) error {
			return errors.New("SSH dial failed: refused")
		}
		err = handleMachineVerify(ctx, db, MachineVerifyPayload{MachineID: id.String()})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SSH dial failed")
	})

	t.Run("success", func(t *testing.T) {
		db := testVerifySQLite(t)
		id := uuid.New()
		host := "127.0.0.1"
		pemData := rsaPEM(t)
		_, err := db.ExecContext(ctx, `INSERT INTO ssh_keys (id, organization_id, name, host, private_key_encrypted, is_active, created_at, updated_at) VALUES (?, ?, 'n', ?, ?, 1, datetime('now'), datetime('now'))`,
			id.String(), uuid.New().String(), host, pemData)
		require.NoError(t, err)
		machineVerifySSHProbe = func(context.Context, string, *cryptossh.ClientConfig) error { return nil }
		err = handleMachineVerify(ctx, db, MachineVerifyPayload{MachineID: id.String()})
		require.NoError(t, err)
	})

	t.Run("context cancelled during probe", func(t *testing.T) {
		db := testVerifySQLite(t)
		id := uuid.New()
		host := "127.0.0.1"
		pemData := rsaPEM(t)
		_, err := db.ExecContext(ctx, `INSERT INTO ssh_keys (id, organization_id, name, host, private_key_encrypted, is_active, created_at, updated_at) VALUES (?, ?, 'n', ?, ?, 1, datetime('now'), datetime('now'))`,
			id.String(), uuid.New().String(), host, pemData)
		require.NoError(t, err)
		pctx, cancel := context.WithCancel(ctx)
		machineVerifySSHProbe = func(c context.Context, _ string, _ *cryptossh.ClientConfig) error {
			<-c.Done()
			return c.Err()
		}
		cancel()
		err = handleMachineVerify(pctx, db, MachineVerifyPayload{MachineID: id.String()})
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func Test_markMachineInactive_invalidID(t *testing.T) {
	db := testVerifySQLite(t)
	markMachineInactive(context.Background(), db, "not-a-uuid")
}

func Test_markMachineInactive_execError(t *testing.T) {
	db := testVerifySQLite(t)
	_ = db.Close()
	markMachineInactive(context.Background(), db, uuid.New().String())
}

func Test_handleMachineVerify_updateFails(t *testing.T) {
	prev := machineVerifySSHProbe
	machineVerifySSHProbe = func(context.Context, string, *cryptossh.ClientConfig) error { return nil }
	t.Cleanup(func() { machineVerifySSHProbe = prev })

	ctx := context.Background()
	db := testVerifySQLite(t)
	_, err := db.ExecContext(ctx, `CREATE TRIGGER trg_block_active BEFORE UPDATE ON ssh_keys FOR EACH ROW WHEN NEW.is_active = 1 BEGIN SELECT RAISE(ABORT, 'blocked'); END`)
	require.NoError(t, err)

	id := uuid.New()
	host := "127.0.0.1"
	pemData := rsaPEM(t)
	_, err = db.ExecContext(ctx, `INSERT INTO ssh_keys (id, organization_id, name, host, private_key_encrypted, is_active, created_at, updated_at) VALUES (?, ?, 'n', ?, ?, 1, datetime('now'), datetime('now'))`,
		id.String(), uuid.New().String(), host, pemData)
	require.NoError(t, err)

	err = handleMachineVerify(ctx, db, MachineVerifyPayload{MachineID: id.String()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update ssh key status")
}
