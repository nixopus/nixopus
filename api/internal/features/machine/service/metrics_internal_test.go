package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

func newMetricsTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqldb.Close() })
	db := bun.NewDB(sqldb, sqlitedialect.New())
	_, err = db.Exec(`CREATE TABLE user_provision_details (
		lxd_container_name TEXT,
		created_at DATETIME,
		ssh_key_id TEXT,
		organization_id TEXT
	)`)
	require.NoError(t, err)
	return db
}

func TestMetricsService_ResolveMachineName_NoRows(t *testing.T) {
	db := newMetricsTestDB(t)
	svc := &MetricsService{db: db}
	_, err := svc.resolveMachineName(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no provisioned machine found")
}

func TestMetricsService_ResolveMachineName_DBError(t *testing.T) {
	db := newMetricsTestDB(t)
	_, err := db.Exec(`DROP TABLE user_provision_details`)
	require.NoError(t, err)
	svc := &MetricsService{db: db}
	_, err = svc.resolveMachineName(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db lookup")
}

func TestMetricsService_ResolveMachineName_EmptyContainerName(t *testing.T) {
	db := newMetricsTestDB(t)
	orgID := uuid.New()
	_, err := db.Exec(`INSERT INTO user_provision_details (lxd_container_name, created_at, organization_id) VALUES (?, ?, ?)`, "", time.Now(), orgID.String())
	require.NoError(t, err)
	svc := &MetricsService{db: db}
	_, err = svc.resolveMachineName(context.Background(), orgID, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "machine name not yet assigned")
}

func TestMetricsService_ResolveMachineName_ByOrg_Success(t *testing.T) {
	db := newMetricsTestDB(t)
	orgID := uuid.New()
	_, err := db.Exec(`INSERT INTO user_provision_details (lxd_container_name, created_at, organization_id) VALUES (?, ?, ?)`, "machine-1", time.Now(), orgID.String())
	require.NoError(t, err)
	svc := &MetricsService{db: db}
	name, err := svc.resolveMachineName(context.Background(), orgID, nil)
	require.NoError(t, err)
	assert.Equal(t, "machine-1", name)
}

func TestMetricsService_ResolveMachineName_ByServerID_Success(t *testing.T) {
	db := newMetricsTestDB(t)
	orgID := uuid.New()
	serverID := uuid.New()
	_, err := db.Exec(`INSERT INTO user_provision_details (lxd_container_name, created_at, organization_id, ssh_key_id) VALUES (?, ?, ?, ?)`, "machine-2", time.Now(), orgID.String(), serverID.String())
	require.NoError(t, err)
	svc := &MetricsService{db: db}
	name, err := svc.resolveMachineName(context.Background(), orgID, &serverID)
	require.NoError(t, err)
	assert.Equal(t, "machine-2", name)
}
