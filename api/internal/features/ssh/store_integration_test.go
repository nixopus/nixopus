package ssh

import (
	"context"
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melbahja/goph"
	"github.com/nixopus/nixopus/api/internal/features/ssh/storage"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidateSSHManagerCache_closesCachedManagers(t *testing.T) {
	resetInvalidateHooksForTest()

	setup := testutils.NewTestSetup()
	withGlobalStore(t, setup)

	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	host, ps, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(ps)
	require.NoError(t, err)

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	key := &types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      org.ID,
		Name:                "srv",
		AuthMethod:          "password",
		IsActive:            true,
		IsDefault:           true,
		Host:                tsp(host),
		User:                tsp("u"),
		Port:                tip(port),
		PasswordEncrypted:   tsp("pw"),
		PrivateKeyEncrypted: nil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	insertSSHKeyRow(t, setup, key)

	mgr, err := GetSSHManagerForServer(setup.Ctx, org.ID, key.ID)
	require.NoError(t, err)

	_, err = mgr.RunCommandWithID("", "ping")
	require.NoError(t, err)

	InvalidateSSHManagerCache(org.ID)

	_, err = GetSSHManagerForServer(setup.Ctx, org.ID, key.ID)
	require.NoError(t, err)

	InvalidateServerManagerCache(key.ID)

	resetInvalidateHooksForTest()
}

func TestInvalidateAllSSHManagerCaches(t *testing.T) {
	setup := testutils.NewTestSetup()
	withGlobalStore(t, setup)

	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	host, ps, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(ps)
	require.NoError(t, err)

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	key := &types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      org.ID,
		Name:                "srv",
		AuthMethod:          "password",
		IsActive:            true,
		IsDefault:           true,
		Host:                tsp(host),
		User:                tsp("u"),
		Port:                tip(port),
		PasswordEncrypted:   tsp("pw"),
		PrivateKeyEncrypted: nil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	insertSSHKeyRow(t, setup, key)

	_, err = GetSSHManagerForServer(setup.Ctx, org.ID, key.ID)
	require.NoError(t, err)

	InvalidateAllSSHManagerCaches()

	_, err = GetSSHManagerForServer(setup.Ctx, org.ID, key.ID)
	require.NoError(t, err)
}

func TestGetSSHManagerForServer_wrongOrg_badCred_badPEM(t *testing.T) {
	setup := testutils.NewTestSetup()
	withGlobalStore(t, setup)

	_, orgA, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	_, orgB, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	key := &types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      orgA.ID,
		Name:                "srv",
		AuthMethod:          "key",
		IsActive:            true,
		IsDefault:           true,
		Host:                tsp("127.0.0.1"),
		User:                tsp("u"),
		Port:                tip(22),
		PasswordEncrypted:   nil,
		PrivateKeyEncrypted: nil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	insertSSHKeyRow(t, setup, key)

	_, err = GetSSHManagerForServer(setup.Ctx, orgB.ID, key.ID)
	require.Error(t, err)

	keyNoCred := &types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      orgA.ID,
		Name:                "srv2",
		AuthMethod:          "key",
		IsActive:            true,
		IsDefault:           false,
		Host:                tsp("127.0.0.1"),
		User:                tsp("u"),
		Port:                tip(22),
		PasswordEncrypted:   nil,
		PrivateKeyEncrypted: nil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	insertSSHKeyRow(t, setup, keyNoCred)

	_, err = GetSSHManagerForServer(setup.Ctx, orgA.ID, keyNoCred.ID)
	require.Error(t, err)

	keyBadPEM := &types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      orgA.ID,
		Name:                "srv3",
		AuthMethod:          "key",
		IsActive:            true,
		IsDefault:           false,
		Host:                tsp("127.0.0.1"),
		User:                tsp("u"),
		Port:                tip(22),
		PasswordEncrypted:   tsp("pw"),
		PrivateKeyEncrypted: tsp("NOT PEM HEADER"),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	insertSSHKeyRow(t, setup, keyBadPEM)

	_, err = GetSSHManagerForServer(setup.Ctx, orgA.ID, keyBadPEM.ID)
	require.Error(t, err)
}

func TestGetSSHManagerForOrganization_promotesActiveAndRunCommand(t *testing.T) {
	setup := testutils.NewTestSetup()
	withGlobalStore(t, setup)

	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	host, ps, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(ps)
	require.NoError(t, err)

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	activeOnly := &types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      org.ID,
		Name:                "promotable",
		AuthMethod:          "password",
		IsActive:            true,
		IsDefault:           false,
		Host:                tsp(host),
		User:                tsp("u"),
		Port:                tip(port),
		PasswordEncrypted:   tsp("pw"),
		PrivateKeyEncrypted: nil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	insertSSHKeyRow(t, setup, activeOnly)

	mgr, err := GetSSHManagerForOrganization(setup.Ctx, org.ID)
	require.NoError(t, err)
	defer mgr.Close()

	store := storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}
	refreshed, err := store.GetSSHKeyByID(activeOnly.ID)
	require.NoError(t, err)
	assert.True(t, refreshed.IsDefault)

	out, err := mgr.RunCommandWithID("", "hello-world")
	require.NoError(t, err)
	assert.Contains(t, out, "hello-world")

	ctx := testSSHOrgContext(setup.Ctx, org.ID)
	mgr2, err := GetSSHManagerFromContext(ctx)
	require.NoError(t, err)
	assert.NotNil(t, mgr2)

	srvCtx := context.WithValue(context.WithValue(setup.Ctx, types.OrganizationIDKey, org.ID.String()), types.ServerIDKey, activeOnly.ID.String())
	mgr3, err := GetSSHManagerFromContext(srvCtx)
	require.NoError(t, err)
	assert.NotNil(t, mgr3)
}

func TestGetSSHManagerFromContext_badServerIDFallsBackToOrg(t *testing.T) {
	setup := testutils.NewTestSetup()
	withGlobalStore(t, setup)

	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	host, ps, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(ps)
	require.NoError(t, err)

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	key := &types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      org.ID,
		Name:                "def",
		AuthMethod:          "password",
		IsActive:            true,
		IsDefault:           true,
		Host:                tsp(host),
		User:                tsp("u"),
		Port:                tip(port),
		PasswordEncrypted:   tsp("pw"),
		PrivateKeyEncrypted: nil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	insertSSHKeyRow(t, setup, key)

	ctx := context.WithValue(context.WithValue(setup.Ctx, types.OrganizationIDKey, org.ID.String()), types.ServerIDKey, "not-a-uuid")
	mgr, err := GetSSHManagerFromContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, mgr)
	mgr.Close()
}

func TestGetSSHManagerForOrganization_noServerConfigured(t *testing.T) {
	setup := testutils.NewTestSetup()
	withGlobalStore(t, setup)

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	_, err = GetSSHManagerForOrganization(setup.Ctx, org.ID)
	require.Error(t, err)
	assert.True(t, IsNoDefaultServerError(err))
}

func TestSSHManager_idleCleanup_evictsUnusedPoolEntry(t *testing.T) {
	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	host, ps, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(ps)
	require.NoError(t, err)

	m := NewSSHManagerForTest(nil, 35*time.Millisecond)
	defer m.Close()

	require.NoError(t, m.AddClient("default", &SSH{
		User:     "u",
		Host:     host,
		Port:     uint(port),
		Password: "pw",
	}))
	_, err = m.Connect()
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	_, err = m.Connect()
	require.NoError(t, err)
}

func TestNewSessionWithRetry_and_RunCommand_manager(t *testing.T) {
	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	host, ps, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(ps)
	require.NoError(t, err)

	m := NewSSHManager()
	defer m.Close()

	require.NoError(t, m.AddClient("default", &SSH{
		User:     "u",
		Host:     host,
		Port:     uint(port),
		Password: "pw",
	}))

	sess, err := m.NewSessionWithRetry("")
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.NoError(t, sess.Close())

	out, err := m.RunCommandWithID("", "abc")
	require.NoError(t, err)
	assert.Contains(t, out, "abc")

	out2, err := m.RunCommand("xyz")
	require.NoError(t, err)
	assert.Contains(t, out2, "xyz")
}

func TestSSHManager_deadPoolConnectionReconnects(t *testing.T) {
	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	host, ps, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(ps)
	require.NoError(t, err)

	m := NewSSHManager()
	defer m.Close()

	require.NoError(t, m.AddClient("default", &SSH{
		User:     "u",
		Host:     host,
		Port:     uint(port),
		Password: "pw",
	}))

	cl1, err := m.ConnectWithID("")
	require.NoError(t, err)
	require.NoError(t, cl1.Close())

	time.Sleep(50 * time.Millisecond)

	cl2, err := m.ConnectWithID("")
	require.NoError(t, err)
	require.NotNil(t, cl2)
}

func TestNewSessionWithRetry_connectionFailure(t *testing.T) {
	m := NewSSHManagerForTest(func(string) (*goph.Client, error) {
		return nil, errors.New("dial failed")
	}, time.Minute)
	defer m.Close()
	require.NoError(t, m.AddClient("default", &SSH{}))

	_, err := m.NewSessionWithRetry("")
	require.Error(t, err)
}
