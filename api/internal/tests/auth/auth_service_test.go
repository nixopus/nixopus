package auth

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	authsvc "github.com/nixopus/nixopus/api/internal/features/auth/service"
	authstor "github.com/nixopus/nixopus/api/internal/features/auth/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestAuthService_BuildBootstrap_happyPath(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	resp, err := svc.BuildBootstrap(setup.Ctx, u, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(resp.Organizations), 1)
	require.Equal(t, org.ID.String(), *resp.ActiveOrganizationID)
	require.False(t, resp.HasServers)
	require.Equal(t, "NOT_STARTED", resp.User.ProvisionStatus)
}

func TestAuthService_BuildBootstrap_sessionActiveOrgOverride(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	other := uuid.New().String()
	resp, err := svc.BuildBootstrap(setup.Ctx, u, &other)
	require.NoError(t, err)
	require.Equal(t, other, *resp.ActiveOrganizationID)
}

func TestAuthService_BuildBootstrap_invalidSessionOrgUUID(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	bad := "not-a-uuid"
	resp, err := svc.BuildBootstrap(setup.Ctx, u, &bad)
	require.NoError(t, err)
	require.False(t, resp.HasServers)
	require.Equal(t, bad, *resp.ActiveOrganizationID)
}

func TestAuthService_BuildBootstrap_listOrgsCanceledContext(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(setup.Ctx)
	cancel()

	resp, err := svc.BuildBootstrap(ctx, u, nil)
	require.NoError(t, err)
	require.Empty(t, resp.Organizations)
}

func TestAuthService_BuildBootstrap_provisioningWithDetails(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	ps := "provisioning"
	u.ProvisionStatus = &ps

	step := types.ProvisionStepInitializing
	details := types.UserProvisionDetails{
		UserID:         u.ID,
		OrganizationID: org.ID,
		Type:           "trial",
		VcpuCount:      1,
		MemoryMB:       512,
		DiskSizeGB:     10,
		Step:           &step,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(&details).Exec(setup.Ctx)
	require.NoError(t, err)

	resp, err := svc.BuildBootstrap(setup.Ctx, u, nil)
	require.NoError(t, err)
	require.NotNil(t, resp.ProvisionID)
	require.NotNil(t, resp.ProvisionStep)
}

func TestAuthService_BuildBootstrap_provisioningStepNil(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	ps := "provisioning"
	u.ProvisionStatus = &ps

	details := types.UserProvisionDetails{
		UserID:         u.ID,
		OrganizationID: org.ID,
		Type:           "trial",
		VcpuCount:      1,
		MemoryMB:       512,
		DiskSizeGB:     10,
		Step:           nil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(&details).Exec(setup.Ctx)
	require.NoError(t, err)

	resp, err := svc.BuildBootstrap(setup.Ctx, u, nil)
	require.NoError(t, err)
	require.NotNil(t, resp.ProvisionID)
	require.Nil(t, resp.ProvisionStep)
}

func TestAuthService_BuildBootstrap_provisioningMissingRow(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	ps := "provisioning"
	u.ProvisionStatus = &ps

	_, err = svc.BuildBootstrap(setup.Ctx, u, nil)
	require.Error(t, err)
}

func TestAuthService_BuildBootstrap_hasServers(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	key := types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "k",
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(&key).Exec(setup.Ctx)
	require.NoError(t, err)

	resp, err := svc.BuildBootstrap(setup.Ctx, u, nil)
	require.NoError(t, err)
	require.True(t, resp.HasServers)
}

func TestAuthService_BuildBootstrap_provisionLabelFromUser(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	ready := "ready"
	u.ProvisionStatus = &ready

	resp, err := svc.BuildBootstrap(setup.Ctx, u, nil)
	require.NoError(t, err)
	require.Equal(t, ready, resp.User.ProvisionStatus)
}

func TestAuthService_BuildBootstrap_provisionStatusEmptyStringUsesPending(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	empty := ""
	u.ProvisionStatus = &empty

	resp, err := svc.BuildBootstrap(setup.Ctx, u, nil)
	require.NoError(t, err)
	require.Equal(t, "pending", resp.User.ProvisionStatus)
}

func TestAuthService_GetAdminRegistered_cacheHitWithDBConfigured(t *testing.T) {
	setup := testutils.NewTestSetup()
	mr := miniredis.RunT(t)
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "redis://"+mr.Addr())
	require.NoError(t, svc.Cache.SetAdminRegistered(setup.Ctx, true))

	got, err := svc.GetAdminRegistered(setup.Ctx)
	require.NoError(t, err)
	require.True(t, got)
}

func TestAuthService_GetAdminRegistered_redisClosedFallsThroughToDB(t *testing.T) {
	setup := testutils.NewTestSetup()
	mr := miniredis.RunT(t)
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "redis://"+mr.Addr())
	mr.Close()

	got, err := svc.GetAdminRegistered(setup.Ctx)
	require.NoError(t, err)
	require.False(t, got)
}

func TestAuthService_GetAdminRegistered_trueAfterSeedCredential(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := logger.NewLogger()
	svc := authsvc.NewAuthService(setup.UserStorage, setup.DB, l, setup.Ctx, "")

	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NoError(t, setup.SeedCredentialAccount(u.ID.String()))

	got, err := svc.GetAdminRegistered(setup.Ctx)
	require.NoError(t, err)
	require.True(t, got)
}

func TestAuthService_GetAdminRegistered_countError(t *testing.T) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5433"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "nixopus"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "nixopus"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "nixopus_test"
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName,
	)
	cfg, err := pgx.ParseConfig(connStr)
	require.NoError(t, err)
	sqldb := stdlib.OpenDB(*cfg)
	deadDB := bun.NewDB(sqldb, pgdialect.New())
	require.NoError(t, deadDB.Ping())
	require.NoError(t, deadDB.Close())

	l := logger.NewLogger()
	st := authstor.UserStorage{DB: deadDB, Ctx: context.Background()}
	svc := authsvc.NewAuthService(&st, deadDB, l, context.Background(), "")

	_, err = svc.GetAdminRegistered(context.Background())
	require.Error(t, err)
}
