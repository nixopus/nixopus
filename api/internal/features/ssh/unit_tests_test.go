package ssh

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melbahja/goph"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tsp(s string) *string { return &s }

func tip(i int) *int { return &i }

func withGlobalStore(t *testing.T, setup *testutils.TestSetup) {
	t.Helper()
	prev := config.GlobalStore
	config.GlobalStore = setup.Store
	t.Cleanup(func() {
		InvalidateAllSSHManagerCaches()
		config.GlobalStore = prev
	})
}

func insertSSHKeyRow(t *testing.T, setup *testutils.TestSetup, row *types.SSHKey) {
	t.Helper()
	_, err := setup.DB.NewInsert().Model(row).Exec(setup.Ctx)
	require.NoError(t, err)
}

func TestIsClosedConnectionError_and_alias(t *testing.T) {
	assert.False(t, IsClosedConnectionError(nil))
	assert.True(t, IsClosedConnectionError(io.EOF))
	assert.True(t, IsClosedConnectionError(fmt.Errorf("broken pipe")))
	assert.True(t, IsClosedConnectionError(fmt.Errorf("connection reset by peer")))
	assert.True(t, IsClosedConnectionError(fmt.Errorf("unexpected packet")))
	assert.False(t, IsClosedConnectionError(fmt.Errorf("other")))
	assert.False(t, isClosedConnectionError(nil))
	assert.True(t, isClosedConnectionError(io.EOF))
}

func TestIsNoDefaultServerError(t *testing.T) {
	assert.False(t, IsNoDefaultServerError(nil))
	assert.True(t, IsNoDefaultServerError(fmt.Errorf("no server configured for organization %s", uuid.New())))
	assert.True(t, IsNoDefaultServerError(errors.New("no default server configured")))
	assert.False(t, IsNoDefaultServerError(errors.New("other")))
}

func TestNewSSHFromConfig(t *testing.T) {
	assert.Nil(t, NewSSHFromConfig(nil))
	cfg := &types.SSHConfig{
		PrivateKey:          "pk",
		Host:                "h",
		ProxyHost:           "p",
		User:                "u",
		Port:                222,
		Password:            "pw",
		PrivateKeyProtected: "prot",
	}
	s := NewSSHFromConfig(cfg)
	require.NotNil(t, s)
	assert.Equal(t, "pk", s.PrivateKey)
	assert.Equal(t, "h", s.Host)
	assert.Equal(t, "p", s.ProxyHost)
	assert.Equal(t, "u", s.User)
	assert.Equal(t, uint(222), s.Port)
	assert.Equal(t, "pw", s.Password)
	assert.Equal(t, "prot", s.PrivateKeyProtected)
}

func TestSSHManager_AddClient_GetClient_SetDefault_ListClients(t *testing.T) {
	m := NewSSHManager()
	require.Error(t, m.AddClient("", &SSH{}))
	require.Error(t, m.AddClient("x", nil))
	require.NoError(t, m.AddClient("default", &SSH{Host: "h"}))

	c, err := m.GetClient("default")
	require.NoError(t, err)
	assert.Equal(t, "h", c.Host)

	_, err = m.GetClient("missing")
	require.Error(t, err)

	require.Error(t, m.SetDefault("nope"))
	require.NoError(t, m.SetDefault("default"))

	ids := m.ListClients()
	require.Len(t, ids, 1)
	assert.Equal(t, "default", ids[0])

	m.Close()
}

func TestSSHManager_GetOrganizationSSH_GetSSHHost_User_Upstream_errors(t *testing.T) {
	m := NewSSHManager()
	require.NoError(t, m.AddClient("default", &SSH{Host: "", User: "", ProxyHost: ""}))

	_, err := m.GetOrganizationSSH()
	require.NoError(t, err)

	_, err = m.GetSSHHost()
	require.Error(t, err)

	m2 := NewSSHManager()
	_, err = m2.GetOrganizationSSH()
	require.Error(t, err)

	require.NoError(t, m2.AddClient("default", &SSH{Host: "srv", User: "", ProxyHost: "jump"}))
	_, err = m2.GetSSHUser()
	require.Error(t, err)

	require.NoError(t, m2.AddClient("default", &SSH{Host: "srv", User: "root", ProxyHost: "jump"}))
	h, err := m2.GetUpstreamHost()
	require.NoError(t, err)
	assert.Equal(t, "jump", h)

	m3 := NewSSHManager()
	require.NoError(t, m3.AddClient("default", &SSH{Host: "only", User: "root"}))
	up, err := m3.GetUpstreamHost()
	require.NoError(t, err)
	assert.Equal(t, "only", up)

	m4 := NewSSHManager()
	require.NoError(t, m4.AddClient("default", &SSH{Host: "", User: "root"}))
	_, err = m4.GetUpstreamHost()
	require.Error(t, err)

	m.Close()
	m2.Close()
	m3.Close()
	m4.Close()
}

func TestInvalidateSSHManagerCache_alwaysFiresHooks(t *testing.T) {
	resetInvalidateHooksForTest()
	var fired uuid.UUID
	RegisterInvalidateHook(func(id uuid.UUID) { fired = id })

	org := uuid.New()
	InvalidateSSHManagerCache(org)
	assert.Equal(t, org, fired)
	resetInvalidateHooksForTest()
}

func TestGetSSHManager_missingGlobalStore(t *testing.T) {
	prev := config.GlobalStore
	config.GlobalStore = nil
	t.Cleanup(func() { config.GlobalStore = prev })

	_, err := GetSSHManagerForOrganization(context.Background(), uuid.New())
	require.Error(t, err)

	_, err = GetSSHManagerForServer(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
}

func TestGetSSHManagerForServer_validationErrors(t *testing.T) {
	setup := testutils.NewTestSetup()
	withGlobalStore(t, setup)

	_, err := GetSSHManagerForServer(setup.Ctx, uuid.Nil, uuid.New())
	require.Error(t, err)

	_, err = GetSSHManagerForServer(setup.Ctx, uuid.New(), uuid.Nil)
	require.Error(t, err)

	_, err = GetSSHManagerForServer(setup.Ctx, uuid.New(), uuid.New())
	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found"))
}

func TestSSH_Connect_helpers_and_TerminalEarlyExit(t *testing.T) {
	_, err := (&SSH{Host: "127.0.0.1", User: ""}).Connect()
	require.Error(t, err)

	_, err = (&SSH{Host: "", User: "u"}).Connect()
	require.Error(t, err)

	_, err = (&SSH{User: "u", Host: "h"}).ConnectWithPrivateKey()
	require.Error(t, err)

	_, err = (&SSH{PrivateKey: "x", User: "u", Host: "127.0.0.1", Port: 65530}).ConnectWithPrivateKey()
	require.Error(t, err)

	_, err = (&SSH{User: "u", Host: "127.0.0.1"}).ConnectWithPassword()
	require.Error(t, err)

	sBad := &SSH{Host: "127.0.0.1", User: "u", PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nBAD\n-----END RSA PRIVATE KEY-----"}
	_, err = sBad.ConnectWithRetry()
	require.Error(t, err)

	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	host, ps, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(ps)
	require.NoError(t, err)

	s := &SSH{
		User:       "u",
		Host:       host,
		Port:       uint(port),
		PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nBAD\n-----END RSA PRIVATE KEY-----",
		Password:   "pw",
	}
	cl, err := s.ConnectWithRetry()
	require.NoError(t, err)
	defer cl.Close()

	out, err := s.RunCommand("echo-x")
	require.NoError(t, err)
	assert.Contains(t, out, "echo-x")

	s.Terminal()

	m := NewSSHManagerForTest(nil, 0)
	defer m.Close()
	assert.NotNil(t, m)
}

func TestSSHManager_ConnectWithID_factoryErrors_and_Borrow(t *testing.T) {
	m := NewSSHManagerForTest(func(string) (*goph.Client, error) {
		return nil, errors.New("boom")
	}, time.Minute)
	defer m.Close()
	require.NoError(t, m.AddClient("default", &SSH{}))

	_, err := m.ConnectWithID("")
	require.Error(t, err)

	m2 := NewSSHManagerForTest(func(string) (*goph.Client, error) {
		return nil, nil
	}, time.Minute)
	defer m2.Close()
	require.NoError(t, m2.AddClient("default", &SSH{}))

	_, err = m2.ConnectWithID("")
	require.Error(t, err)

	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	m3 := NewSSHManagerForTest(func(string) (*goph.Client, error) {
		cl, err := dialPasswordGoph(t, addr, "u", "pw")
		require.NoError(t, err)
		return cl, nil
	}, time.Minute)
	defer m3.Close()
	require.NoError(t, m3.AddClient("default", &SSH{}))

	cl, rel, err := m3.Borrow("")
	require.NoError(t, err)
	require.NotNil(t, cl)
	rel()

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := m3.ConnectWithID("")
			assert.NoError(t, err)
			assert.NotNil(t, c)
		}()
	}
	wg.Wait()

	m3.CloseConnection("")
	m3.Close()
	m3.Close()
}

func Test_sendKeepalive_closedClient(t *testing.T) {
	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	cl, err := dialPasswordGoph(t, addr, "u", "pw")
	require.NoError(t, err)
	require.NoError(t, cl.Close())

	assert.False(t, sendKeepalive(cl, time.Second))
}

func Test_StartKeepalive_nilClient_noop(t *testing.T) {
	StartKeepalive(nil, time.Millisecond, 4, nil)
}

func Test_StartKeepalive_stopImmediate(t *testing.T) {
	addr, shutdown := startPasswordEchoSSHServer(t, "u", "pw")
	defer shutdown()

	cl, err := dialPasswordGoph(t, addr, "u", "pw")
	require.NoError(t, err)
	defer cl.Close()

	stop := make(chan struct{})
	close(stop)
	StartKeepalive(cl, time.Millisecond, 4, stop)
	time.Sleep(30 * time.Millisecond)
}
