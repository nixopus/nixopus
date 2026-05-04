package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func sqliteAuthDB(t *testing.T) (*bun.DB, context.Context) {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file:memauth"+uuid.New().String()+"?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	ctx := context.Background()

	db.RegisterModel((*types.Organization)(nil))
	db.RegisterModel((*types.Member)(nil))
	db.RegisterModel((*types.OrganizationUsers)(nil))
	db.RegisterModel((*types.User)(nil))
	db.RegisterModel((*Account)(nil))
	db.RegisterModel((*types.SSHKey)(nil))
	db.RegisterModel((*types.UserProvisionDetails)(nil))

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
		`CREATE TABLE account (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			account_id TEXT,
			provider_id TEXT NOT NULL,
			password TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
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
			last_used_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			deleted_at TEXT
		)`,
		`CREATE TABLE user_provision_details (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			organization_id TEXT NOT NULL,
			server_id TEXT,
			guest_ip TEXT,
			lxd_container_name TEXT,
			ssh_key_id TEXT,
			type TEXT NOT NULL DEFAULT 'trial',
			subdomain TEXT,
			domain TEXT,
			vcpu_count INTEGER NOT NULL DEFAULT 0,
			memory_mb INTEGER NOT NULL DEFAULT 0,
			disk_size_gb INTEGER NOT NULL DEFAULT 0,
			step TEXT,
			error TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
	} {
		_, err = db.ExecContext(ctx, stmt)
		require.NoError(t, err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db, ctx
}

func seedUserOrgMember(t *testing.T, db *bun.DB, ctx context.Context) (*types.User, *types.Organization) {
	t.Helper()
	userID := uuid.New()
	orgID := uuid.New()
	memberID := uuid.New()
	now := time.Now().UTC()

	u := &types.User{
		ID:            userID,
		Name:          "Test",
		Email:         "u+" + userID.String()[:8] + "@test.local",
		EmailVerified: true,
		IsOnboarded:   true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	org := &types.Organization{
		ID:        orgID,
		Name:      "Org",
		Slug:      "slug-" + orgID.String()[:8],
		CreatedAt: now,
	}
	m := &types.Member{
		ID:             memberID,
		OrganizationID: orgID,
		UserID:         userID,
		Role:           types.RoleMember,
		CreatedAt:      now,
	}

	_, err := db.NewInsert().Model(u).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(org).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(m).Exec(ctx)
	require.NoError(t, err)

	return u, org
}

func TestUserStorage_BeginTx(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	tx, err := us.BeginTx()
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
}

func TestUserStorage_WithTx_FindUserByEmail(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	base := &UserStorage{DB: db, Ctx: ctx}

	u, _ := seedUserOrgMember(t, db, ctx)

	tx, err := base.BeginTx()
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })

	scoped := base.WithTx(tx)
	got, err := scoped.FindUserByEmail(u.Email)
	require.NoError(t, err)
	require.Equal(t, u.ID, got.ID)

	require.NoError(t, tx.Rollback())

	_, err = scoped.FindUserByEmail(u.Email)
	require.Error(t, err)
}

func TestUserStorage_FindUserByEmail(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	u, _ := seedUserOrgMember(t, db, ctx)

	got, err := us.FindUserByEmail(u.Email)
	require.NoError(t, err)
	require.Equal(t, u.ID, got.ID)
	require.Equal(t, u.Email, got.Email)

	_, err = us.FindUserByEmail("no-one-" + uuid.New().String() + "@example.test")
	require.Error(t, err)
}

func TestUserStorage_CountAccountsWithPassword(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	n, err := us.CountAccountsWithPassword(ctx)
	require.NoError(t, err)
	require.Zero(t, n)

	user, _ := seedUserOrgMember(t, db, ctx)
	now := time.Now().UTC()
	acct := &Account{
		ID:         uuid.New(),
		UserID:     user.ID,
		AccountID:  user.ID.String(),
		ProviderID: "credential",
		Password:   ptr("hash"),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_, err = db.NewInsert().Model(acct).Exec(ctx)
	require.NoError(t, err)

	n, err = us.CountAccountsWithPassword(ctx)
	require.NoError(t, err)
	require.Positive(t, n)
}

func TestUserStorage_ListBootstrapOrganizations(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	user, org := seedUserOrgMember(t, db, ctx)

	rows, err := us.ListBootstrapOrganizations(ctx, user.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rows), 1)
	require.Equal(t, org.ID, rows[0].ID)
}

func TestUserStorage_OrgHasSSHKeys(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	user, org := seedUserOrgMember(t, db, ctx)
	_ = user

	ok, err := us.OrgHasSSHKeys(ctx, org.ID)
	require.NoError(t, err)
	require.False(t, ok)

	now := time.Now().UTC()
	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "k",
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	_, err = db.NewInsert().Model(key).Exec(ctx)
	require.NoError(t, err)

	ok, err = us.OrgHasSSHKeys(ctx, org.ID)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestUserStorage_GetLatestUserProvisionDetails(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	user, org := seedUserOrgMember(t, db, ctx)

	_, err := us.GetLatestUserProvisionDetails(ctx, user.ID)
	require.Error(t, err)

	step := types.ProvisionStepInitializing
	details := &types.UserProvisionDetails{
		ID:             uuid.New(),
		UserID:         user.ID,
		OrganizationID: org.ID,
		Type:           "trial",
		VcpuCount:      1,
		MemoryMB:       512,
		DiskSizeGB:     10,
		Step:           &step,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	_, err = db.NewInsert().Model(details).Exec(ctx)
	require.NoError(t, err)

	got, err := us.GetLatestUserProvisionDetails(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, details.ID, got.ID)
	require.NotNil(t, got.Step)
}

func TestUserStorage_GetLatestUserProvisionDetails_rowWithoutStep(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	user, org := seedUserOrgMember(t, db, ctx)

	details := &types.UserProvisionDetails{
		ID:             uuid.New(),
		UserID:         user.ID,
		OrganizationID: org.ID,
		Type:           "trial",
		VcpuCount:      1,
		MemoryMB:       512,
		DiskSizeGB:     10,
		Step:           nil,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	_, err := db.NewInsert().Model(details).Exec(ctx)
	require.NoError(t, err)

	got, err := us.GetLatestUserProvisionDetails(ctx, user.ID)
	require.NoError(t, err)
	require.Nil(t, got.Step)
}

func TestUserStorage_ListBootstrapOrganizations_canceledContext(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	user, _ := seedUserOrgMember(t, db, ctx)

	ctx2, cancel := context.WithCancel(ctx)
	cancel()

	_, err := us.ListBootstrapOrganizations(ctx2, user.ID)
	require.Error(t, err)
}

func TestUserStorage_CountAccountsWithPassword_canceledContext(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	ctx2, cancel := context.WithCancel(ctx)
	cancel()

	_, err := us.CountAccountsWithPassword(ctx2)
	require.Error(t, err)
}

func TestUserStorage_OrgHasSSHKeys_canceledContext(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	ctx2, cancel := context.WithCancel(ctx)
	cancel()

	_, err := us.OrgHasSSHKeys(ctx2, uuid.New())
	require.Error(t, err)
}

func TestUserStorage_GetLatestUserProvisionDetails_canceledContext(t *testing.T) {
	t.Parallel()
	db, ctx := sqliteAuthDB(t)
	us := &UserStorage{DB: db, Ctx: ctx}

	ctx2, cancel := context.WithCancel(ctx)
	cancel()

	_, err := us.GetLatestUserProvisionDetails(ctx2, uuid.New())
	require.Error(t, err)
}

func ptr(s string) *string { return &s }
