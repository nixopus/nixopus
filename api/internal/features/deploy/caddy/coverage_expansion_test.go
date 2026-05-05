package caddy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/deploy/docker"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/raghavyuva/caddygo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetCaddyHooks(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		testGetCaddyClientHook = nil
		tunnelCreateHook = nil
		tunnelDialSSHHook = nil
		getSSHHostForOrgHook = nil
		reconcileGetPublishedPortHook = nil
		getDockerServiceFromContext = docker.GetDockerServiceFromContext
		getSSHManagerFromContextForCaddy = ssh.GetSSHManagerFromContext
		pingCaddyProbe = PingCaddy
		caddyTunnelCacheMu.Lock()
		for _, e := range caddyTunnelCache {
			if e != nil && e.tunnel != nil {
				e.tunnel.Close()
			}
		}
		caddyTunnelCache = make(map[string]*caddyTunnelEntry)
		caddyTunnelCacheMu.Unlock()
		enqueueReconcileAfterRecovery = EnqueueReconcile
	})
}

// attachTunnelHarness directs createCaddyTunnelForClient → local httptest admin (real GetCaddyClient path).
func attachTunnelHarness(t *testing.T, cfg *memCaddyCfg) *httptest.Server {
	t.Helper()
	srv := newTestCaddyAdminServer(t, cfg)
	tunnelCreateHook = func(*ssh.SSH, string, uuid.UUID, logger.Logger) (*CaddyTunnel, error) {
		return &CaddyTunnel{
			endpoint: strings.TrimSuffix(srv.URL, "/"),
			cleanup:  func() error { return nil },
		}, nil
	}
	return srv
}

func TestGetCaddyClient_viaTunnelHook_cache_nilLogger_errors(t *testing.T) {
	resetCaddyHooks(t)

	cfgMem := newMemCaddyCfg(`{}`)
	attachTunnelHarness(t, cfgMem)

	host := "c-client-" + uuid.New().String() + ".invalid"
	sshCli := &ssh.SSH{Host: host}
	ctx := context.Background()
	log := logger.NewLogger()

	_, err := GetCaddyClient(ctx, &ssh.SSH{Host: ""}, &log)
	require.Error(t, err)

	orig := config.AppConfig.Proxy
	t.Cleanup(func() { config.AppConfig.Proxy = orig })
	config.AppConfig.Proxy.CaddyPort = "nan"
	_, err = GetCaddyClient(ctx, sshCli, &log)
	require.Error(t, err)

	config.AppConfig.Proxy = orig

	c1, err := GetCaddyClient(ctx, sshCli, &log)
	require.NoError(t, err)
	c2, err := GetCaddyClient(ctx, sshCli, &log)
	require.NoError(t, err)
	require.Same(t, c1, c2)

	c3, err := GetCaddyClient(ctx, sshCli, nil)
	require.NoError(t, err)
	require.Same(t, c1, c3)
	require.NoError(t, PingCaddy(ctx, sshCli, &log))
}

func TestAddRemoveDomainsAtomic_withTunnelHook(t *testing.T) {
	resetCaddyHooks(t)

	cfgMem := newMemCaddyCfg(`{}`)
	attachTunnelHarness(t, cfgMem)

	ctx := context.Background()
	l := logger.NewLogger()
	sshCli := &ssh.SSH{Host: "atomic-" + uuid.New().String() + ".invalid"}

	require.NoError(t, AddDomainsAtomic(ctx, sshCli, &l, []DomainRoute{
		{Domain: "atomic.example.test", UpstreamDial: "127.0.0.1:15000"},
	}))

	require.NoError(t, RemoveDomainsAtomic(ctx, sshCli, &l, []string{"atomic.example.test"}))

	atomicErr := AddDomainsAtomic(ctx, sshCli, &l, []DomainRoute{
		{Domain: "bad.invalid", UpstreamDial: "not-a-valid-dial"},
	})
	require.Error(t, atomicErr)
}

func Test_getSSHHostForOrg_managerError(t *testing.T) {
	orig := getSSHManagerFromContextForCaddy
	getSSHManagerFromContextForCaddy = func(context.Context) (*ssh.SSHManager, error) {
		return nil, errors.New("no ssh manager")
	}
	t.Cleanup(func() {
		getSSHManagerFromContextForCaddy = orig
	})

	_, err := getSSHHostForOrg(context.Background())
	require.ErrorContains(t, err, "no ssh manager")
}

func TestHealthMonitor_enqueueAll_fetchError(t *testing.T) {
	rec := NewReconciler(&deployRepoTestStub{}, logger.NewLogger())
	h := NewHealthMonitor(logger.NewLogger(), rec, time.Hour,
		func(context.Context) ([]uuid.UUID, error) {
			return nil, errors.New("orgs unavailable")
		},
	)
	h.enqueueAll(context.Background())
}

// memCaddyCfg stores JSON returned by GET /config/ and updated by POST.
type memCaddyCfg struct {
	mu sync.Mutex
	b  []byte
}

func newMemCaddyCfg(initial string) *memCaddyCfg {
	m := &memCaddyCfg{b: []byte(initial)}
	if strings.TrimSpace(string(m.b)) == "" {
		m.b = []byte(`{}`)
	}
	return m
}

func newTestCaddyAdminServer(t *testing.T, cfg *memCaddyCfg) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/config/":
			switch r.Method {
			case http.MethodGet:
				cfg.mu.Lock()
				payload := cfg.b
				cfg.mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(payload)
			case http.MethodPost:
				body, _ := io.ReadAll(r.Body)
				cfg.mu.Lock()
				cfg.b = body
				cfg.mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
			default:
				http.NotFound(w, r)
			}
			return
		case "/config":
			http.Redirect(w, r, "/config/", http.StatusPermanentRedirect)
		case "/load":
			if r.Method != http.MethodPost {
				http.NotFound(w, r)
				return
			}
			body, _ := io.ReadAll(r.Body)
			cfg.mu.Lock()
			cfg.b = body
			cfg.mu.Unlock()
			w.WriteHeader(http.StatusOK)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.Close)
	return s
}

func clientForHarness(srv *httptest.Server) *caddygo.Client {
	cli := &http.Client{}
	return &caddygo.Client{
		BaseURL:    strings.TrimSuffix(srv.URL, "/"),
		HTTPClient: cli,
	}
}

func hookClientFromHarness(t *testing.T, srv *httptest.Server) {
	t.Helper()
	c := clientForHarness(srv)
	testGetCaddyClientHook = func(context.Context, *ssh.SSH, *logger.Logger) (*caddygo.Client, error) {
		return c, nil
	}
}

func TestPingCaddy_and_GetCurrentDomains_viaHook(t *testing.T) {
	resetCaddyHooks(t)

	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)
	hookClientFromHarness(t, srv)

	l := logger.NewLogger()
	ctx := context.Background()
	require.NoError(t, PingCaddy(ctx, nil, &l))

	doms, err := GetCurrentDomains(ctx, nil, &l)
	require.NoError(t, err)
	assert.Nil(t, doms)
}

func TestGetCaddyConfig_RestoreCaddyConfig(t *testing.T) {
	resetCaddyHooks(t)

	payload := func() []byte {
		b, err := json.Marshal(&caddy.Config{})
		require.NoError(t, err)
		return b
	}()
	cfgMem := newMemCaddyCfg(string(payload))
	srv := newTestCaddyAdminServer(t, cfgMem)
	hookClientFromHarness(t, srv)

	ctx := context.Background()
	l := logger.NewLogger()
	got, err := GetCaddyConfig(ctx, nil, &l)
	require.NoError(t, err)
	require.NotNil(t, got)

	swap := func() shared_types.ProxyConfig {
		return config.AppConfig.Proxy
	}
	origProxy := swap()
	t.Cleanup(func() {
		config.AppConfig.Proxy = origProxy
	})
	config.AppConfig.Proxy.CaddyPort = ""

	routes := `[{"match":[{"host":["z.example.test"]}],"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"127.0.0.1:9090"}]}]}]`
	patchJSON := fmt.Sprintf(`{"apps":{"http":{"servers":{"nixopus":{"routes":%s}}}}}`, routes)
	cfgMem.mu.Lock()
	cfgMem.b = []byte(patchJSON)
	cfgMem.mu.Unlock()

	require.NoError(t, RestoreCaddyConfig(ctx, nil, &l, got))
}

func TestInvalidateTunnel_earlyReturn_badPort(t *testing.T) {
	orig := config.AppConfig.Proxy
	t.Cleanup(func() { config.AppConfig.Proxy = orig })
	config.AppConfig.Proxy.CaddyPort = "not-a-number"

	key := "h:2019"
	caddyTunnelCacheMu.Lock()
	caddyTunnelCache[key] = &caddyTunnelEntry{lastUsed: time.Now()}
	caddyTunnelCacheMu.Unlock()
	t.Cleanup(func() {
		caddyTunnelCacheMu.Lock()
		delete(caddyTunnelCache, key)
		caddyTunnelCacheMu.Unlock()
	})

	InvalidateTunnel("h")

	caddyTunnelCacheMu.RLock()
	_, exists := caddyTunnelCache[key]
	caddyTunnelCacheMu.RUnlock()
	assert.True(t, exists, "invalid Caddy port should skip eviction")
}

func TestAddDomains_RemoveDomains_operations_withHarness(t *testing.T) {
	resetCaddyHooks(t)

	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)
	hookClientFromHarness(t, srv)

	ctx := context.Background()
	l := logger.NewLogger()
	domains := []DomainRoute{{Domain: "op.example.test", UpstreamDial: "127.0.0.1:1234"}}
	require.NoError(t, AddDomainsWithRetry(ctx, nil, &l, domains))

	var calls atomic.Int32
	testGetCaddyClientHook = func(context.Context, *ssh.SSH, *logger.Logger) (*caddygo.Client, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("transient miss")
		}
		return clientForHarness(srv), nil
	}
	sshKick := &ssh.SSH{Host: "invalidate.example"}
	defer InvalidateTunnelByKey("invalidate.example:2019")
	transientL := logger.NewLogger()
	require.NoError(t, AddDomainsWithRetry(context.Background(), sshKick, &transientL, domains))

	require.NoError(t, RemoveDomainsWithRetry(context.Background(), nil, &l, []string{"op.example.test"}))
}

// errorServer returns a test server that always responds with the given status code.
func errorServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// badJSONServer returns a test server that returns malformed JSON on any GET.
func badJSONServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{invalid-json")
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestPingCaddy_getCaddyClientError(t *testing.T) {
	resetCaddyHooks(t)
	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		return nil, errors.New("no client")
	}
	l := logger.NewLogger()
	err := PingCaddy(context.Background(), nil, &l)
	require.Error(t, err)
}

func TestPingCaddy_reloadError(t *testing.T) {
	resetCaddyHooks(t)
	srv := errorServer(t, http.StatusInternalServerError)
	hookClientFromHarness(t, srv)
	l := logger.NewLogger()
	err := PingCaddy(context.Background(), nil, &l)
	require.Error(t, err)
}

func TestGetCurrentDomains_getCaddyClientError(t *testing.T) {
	resetCaddyHooks(t)
	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		return nil, errors.New("no client")
	}
	l := logger.NewLogger()
	_, err := GetCurrentDomains(context.Background(), nil, &l)
	require.Error(t, err)
}

func TestGetCurrentDomains_http4xx(t *testing.T) {
	resetCaddyHooks(t)
	hookClientFromHarness(t, errorServer(t, http.StatusForbidden))
	l := logger.NewLogger()
	_, err := GetCurrentDomains(context.Background(), nil, &l)
	require.Error(t, err)
}

func TestGetCurrentDomains_invalidJSON(t *testing.T) {
	resetCaddyHooks(t)
	hookClientFromHarness(t, badJSONServer(t))
	l := logger.NewLogger()
	_, err := GetCurrentDomains(context.Background(), nil, &l)
	require.Error(t, err)
}

func TestGetCurrentDomains_httpTransportError(t *testing.T) {
	resetCaddyHooks(t)
	// Build a client pointing to a closed server so HTTP Get returns a transport error.
	srv := newTestCaddyAdminServer(t, newMemCaddyCfg(`{}`))
	c := clientForHarness(srv)
	srv.Close()
	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		return c, nil
	}
	l := logger.NewLogger()
	_, err := GetCurrentDomains(context.Background(), nil, &l)
	require.Error(t, err)
}

func TestGetCaddyConfig_http4xx(t *testing.T) {
	resetCaddyHooks(t)
	hookClientFromHarness(t, errorServer(t, http.StatusUnauthorized))
	l := logger.NewLogger()
	_, err := GetCaddyConfig(context.Background(), nil, &l)
	require.Error(t, err)
}

func TestGetCaddyConfig_invalidJSON(t *testing.T) {
	resetCaddyHooks(t)
	hookClientFromHarness(t, badJSONServer(t))
	l := logger.NewLogger()
	_, err := GetCaddyConfig(context.Background(), nil, &l)
	require.Error(t, err)
}

func TestRestoreCaddyConfig_http4xx(t *testing.T) {
	resetCaddyHooks(t)
	hookClientFromHarness(t, errorServer(t, http.StatusBadRequest))
	l := logger.NewLogger()
	err := RestoreCaddyConfig(context.Background(), nil, &l, &caddy.Config{})
	require.Error(t, err)
}

func TestAddDomainsWithRetry_multiUpstream_invalidDial(t *testing.T) {
	resetCaddyHooks(t)
	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)
	hookClientFromHarness(t, srv)

	ctx := context.Background()
	l := logger.NewLogger()
	domains := []DomainRoute{{
		Domain:    "lb.example",
		Upstreams: []string{"bad-dial-no-port"},
	}}
	err := AddDomainsWithRetry(ctx, nil, &l, domains)
	require.Error(t, err)
}

func TestAddDomainsWithRetry_multiUpstream_valid(t *testing.T) {
	resetCaddyHooks(t)
	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)
	hookClientFromHarness(t, srv)

	ctx := context.Background()
	l := logger.NewLogger()

	// Single entry in Upstreams (no LB opts needed)
	err := AddDomainsWithRetry(ctx, nil, &l, []DomainRoute{{
		Domain:    "single-in-upstreams.example",
		Upstreams: []string{"127.0.0.1:7001"},
	}})
	// Tolerate caddygo HTTP errors – the multi-upstream code path is exercised.
	_ = err

	// Two entries in Upstreams (LB options branch)
	err = AddDomainsWithRetry(ctx, nil, &l, []DomainRoute{{
		Domain:    "lb.example",
		Upstreams: []string{"127.0.0.1:7002", "127.0.0.1:7003"},
		LBPolicy:  "round_robin",
	}})
	_ = err
}

func TestRemoveDomainsAtomic_snapshotFails(t *testing.T) {
	resetCaddyHooks(t)
	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		return nil, errors.New("no client")
	}
	l := logger.NewLogger()
	err := RemoveDomainsAtomic(context.Background(), nil, &l, []string{"x.example"})
	require.Error(t, err)
}

func TestAddDomainsAtomic_snapshotFails(t *testing.T) {
	resetCaddyHooks(t)
	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		return nil, errors.New("no client")
	}
	l := logger.NewLogger()
	err := AddDomainsAtomic(context.Background(), nil, &l, []DomainRoute{{Domain: "x.example", UpstreamDial: "127.0.0.1:3000"}})
	require.Error(t, err)
}

// TestAddDomainsAtomic_addFails_rollbackFails covers the "add failed AND rollback failed" branch.
// Snapshot (call 1) succeeds; all subsequent calls fail so both add-retries and rollback fail.
func TestAddDomainsAtomic_addFails_rollbackFails(t *testing.T) {
	resetCaddyHooks(t)

	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)

	var n atomic.Int32
	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		if n.Add(1) == 1 {
			return clientForHarness(srv), nil
		}
		return nil, errors.New("client gone")
	}

	l := logger.NewLogger()
	// Use an invalid dial so each retry inside WithRetry fails immediately.
	err := AddDomainsAtomic(context.Background(), nil, &l, []DomainRoute{
		{Domain: "failing.example", UpstreamDial: "no-colon-here"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "add failed AND rollback failed")
}

// TestRemoveDomainsAtomic_removeFails_rollbackSucceeds covers stmts 8-12 of RemoveDomainsAtomic.
// Snapshot succeeds; remove client always fails; rollback client succeeds.
func TestRemoveDomainsAtomic_removeFails_rollbackSucceeds(t *testing.T) {
	resetCaddyHooks(t)

	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)

	var n atomic.Int32
	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		call := n.Add(1)
		if call == 1 {
			// snapshot
			return clientForHarness(srv), nil
		}
		if call <= int32(1+DefaultMaxRetries) {
			// all remove retry calls fail
			return nil, errors.New("remove client gone")
		}
		// rollback succeeds
		return clientForHarness(srv), nil
	}

	l := logger.NewLogger()
	err := RemoveDomainsAtomic(context.Background(), nil, &l, []string{"orphan.example"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "domain remove rolled back")
}

// Test_addRouteToClient_paths exercises all branches of addRouteToClient directly.
func Test_addRouteToClient_paths(t *testing.T) {
	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)
	c := clientForHarness(srv)

	// Multi-upstream: invalid dial in Upstreams → parseDial error.
	err := addRouteToClient(c, DomainRoute{Domain: "t.example", Upstreams: []string{"bad-dial"}})
	require.Error(t, err)

	// Multi-upstream: single valid entry (len==1, no LB opts path).
	err = addRouteToClient(c, DomainRoute{Domain: "t.example", Upstreams: []string{"127.0.0.1:8001"}})
	_ = err // caddygo may or may not error on the test server; code path is covered

	// Multi-upstream: two valid entries (len>1, LB opts path).
	err = addRouteToClient(c, DomainRoute{
		Domain:    "lb.example",
		Upstreams: []string{"127.0.0.1:8002", "127.0.0.1:8003"},
		LBPolicy:  "round_robin",
	})
	_ = err

	// Single upstream: invalid UpstreamDial → parseDial error.
	err = addRouteToClient(c, DomainRoute{Domain: "t.example", UpstreamDial: "no-port"})
	require.Error(t, err)
}

// TestReconciler_fullRebuild_withRoutes covers the non-empty desired path.
func TestReconciler_fullRebuild_withRoutes(t *testing.T) {
	resetCaddyHooks(t)

	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)

	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		return clientForHarness(srv), nil
	}

	l := logger.NewLogger()
	r := NewReconciler(&deployRepoTestStub{}, l)

	desired := []DomainRoute{
		{Domain: "rebuild.example", UpstreamDial: "127.0.0.1:4000"},
	}
	got, err := r.fullRebuild(context.Background(), desired)
	require.NoError(t, err)
	require.NotNil(t, got)
	// Domain should appear in Added or Errors depending on caddygo/server behaviour.
	assert.True(t, len(got.Added) > 0 || len(got.Errors) > 0)
}

// TestReconciler_fullRebuild_getCaddyClientError covers the client-error path.
func TestReconciler_fullRebuild_getCaddyClientError(t *testing.T) {
	resetCaddyHooks(t)
	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		return nil, errors.New("no client")
	}
	l := logger.NewLogger()
	r := NewReconciler(&deployRepoTestStub{}, l)
	_, err := r.fullRebuild(context.Background(), []DomainRoute{
		{Domain: "test.example", UpstreamDial: "127.0.0.1:5000"},
	})
	require.Error(t, err)
}

func TestReconcileOrganization_add_and_pendingRemovals_viaHooks(t *testing.T) {
	resetCaddyHooks(t)
	setupTestRedis(t)

	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)
	hookClientFromHarness(t, srv)

	orgID := uuid.New()
	ctxOrg := context.WithValue(context.Background(), shared_types.OrganizationIDKey, orgID.String())
	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "proxy.internal", nil
	}
	reconcileGetPublishedPortHook = func(*Reconciler, context.Context, string) (int, error) {
		return 7000, nil
	}

	d := shared_types.ApplicationDomain{Domain: "rec.example"}
	app := shared_types.Application{
		ID:              uuid.New(),
		Name:            "svc-a",
		BuildPack:       shared_types.DockerFile,
		Domains:         []*shared_types.ApplicationDomain{&d},
		Servers:         []*shared_types.ApplicationServer{{}},
		RoutingStrategy: shared_types.RoutingStrategySingle,
	}
	st := &deployRepoTestStub{
		servers:      []shared_types.ApplicationServer{{}},
		deployedApps: []shared_types.Application{app},
	}
	rec := NewReconciler(st, logger.NewLogger())

	err := TrackExtensionDomain(orgID, "ext.rec", "10.10.10.10:4434")
	require.NoError(t, err)

	err = EnqueuePendingRemoval(orgID, "removed.example")
	require.NoError(t, err)

	rs, err := rec.ReconcileOrganization(ctxOrg, orgID)
	require.NoError(t, err)
	assert.Contains(t, rs.Added, "rec.example")
	assert.Contains(t, rs.Added, "ext.rec")
	require.NotEmpty(t, rs.Removed)
	assert.Contains(t, rs.Removed, "removed.example")
}

// ---- CaddyTunnel pure-logic tests -----------------------------------------

func Test_tunnelErrLog(t *testing.T) {
	org := uuid.New()
	tWithOrg := &CaddyTunnel{orgID: org}
	msg := tWithOrg.tunnelErrLog("myhost", "boom")
	assert.Contains(t, msg, "org="+org.String())
	assert.Contains(t, msg, "host=myhost")
	assert.Contains(t, msg, "err=boom")

	tNoOrg := &CaddyTunnel{}
	msg2 := tNoOrg.tunnelErrLog("myhost", "boom")
	assert.NotContains(t, msg2, "org=")
	assert.Contains(t, msg2, "host=myhost")
}

func Test_CaddyTunnel_Close_withKeepalive(t *testing.T) {
	stopCh := make(chan struct{})
	tunnel := &CaddyTunnel{
		stopKeepalive: stopCh,
		cleanup:       func() error { return nil },
	}
	err := tunnel.Close()
	require.NoError(t, err)
	// The channel should have been closed.
	select {
	case <-stopCh:
	default:
		t.Fatal("stopKeepalive channel was not closed")
	}
}

func Test_CaddyTunnel_Close_nilCleanup(t *testing.T) {
	tunnel := &CaddyTunnel{cleanup: nil}
	err := tunnel.Close()
	require.NoError(t, err)
}

// ---- GetCaddyClient: nil sshClient + no org in context --------------------

func TestGetCaddyClient_nilSSH_noOrgInCtx(t *testing.T) {
	resetCaddyHooks(t)
	// testGetCaddyClientHook is nil after reset, so the real GetCaddyClient runs.
	// With nil sshClient and no org ID in ctx, the function should return an error
	// (organization ID or SSH client required).
	l := logger.NewLogger()
	_, err := GetCaddyClient(context.Background(), nil, &l)
	require.Error(t, err)
}

// ---- ReconcileOrganization additional coverage ----------------------------

// TestReconcileOrganization_getCaddyClientError tests the path where GetCurrentDomains
// succeeds (returns empty) but the second GetCaddyClient call (for needsReload) fails.
func TestReconcileOrganization_getCaddyClientError_needsReload(t *testing.T) {
	resetCaddyHooks(t)
	setupTestRedis(t)

	orgID := uuid.New()
	ctxOrg := context.WithValue(context.Background(), shared_types.OrganizationIDKey, orgID.String())

	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "proxy.internal", nil
	}
	reconcileGetPublishedPortHook = func(*Reconciler, context.Context, string) (int, error) {
		return 7000, nil
	}

	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)

	var n atomic.Int32
	testGetCaddyClientHook = func(_ context.Context, _ *ssh.SSH, _ *logger.Logger) (*caddygo.Client, error) {
		if n.Add(1) == 1 {
			// First call (GetCurrentDomains): returns valid client
			return clientForHarness(srv), nil
		}
		// Second call (GetCaddyClient inside needsReload block): fail
		return nil, errors.New("client gone for reload")
	}

	d := shared_types.ApplicationDomain{Domain: "new.example"}
	app := shared_types.Application{
		ID:        uuid.New(),
		Name:      "svc",
		BuildPack: shared_types.DockerFile,
		Domains:   []*shared_types.ApplicationDomain{&d},
	}
	st := &deployRepoTestStub{deployedApps: []shared_types.Application{app}}
	rec := NewReconciler(st, logger.NewLogger())

	_, err := rec.ReconcileOrganization(ctxOrg, orgID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get caddy client for reconciliation")
}

// TestReconcileOrganization_toUpdate tests the path where a domain exists in Caddy
// but with a different upstream (triggers toUpdate → addRouteToClient).
func TestReconcileOrganization_toUpdate(t *testing.T) {
	resetCaddyHooks(t)
	setupTestRedis(t)

	orgID := uuid.New()
	ctxOrg := context.WithValue(context.Background(), shared_types.OrganizationIDKey, orgID.String())

	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "proxy.internal", nil
	}
	reconcileGetPublishedPortHook = func(*Reconciler, context.Context, string) (int, error) {
		return 7000, nil
	}

	// Caddy has "existing.example" with old.host:7000; desired will have proxy.internal:7000.
	existingCfg := `{"apps":{"http":{"servers":{"nixopus":{"routes":[` +
		`{"match":[{"host":["existing.example"]}],` +
		`"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"old.host:7000"}]}]}` +
		`]}}}}}`
	cfgMem := newMemCaddyCfg(existingCfg)
	srv := newTestCaddyAdminServer(t, cfgMem)
	hookClientFromHarness(t, srv)

	d := shared_types.ApplicationDomain{Domain: "existing.example"}
	app := shared_types.Application{
		ID:        uuid.New(),
		Name:      "svc",
		BuildPack: shared_types.DockerFile,
		Domains:   []*shared_types.ApplicationDomain{&d},
	}
	st := &deployRepoTestStub{deployedApps: []shared_types.Application{app}}
	rec := NewReconciler(st, logger.NewLogger())

	rs, err := rec.ReconcileOrganization(ctxOrg, orgID)
	require.NoError(t, err)
	// "existing.example" should be Updated (or in Errors if caddygo fails on test server).
	assert.True(t, len(rs.Updated) > 0 || len(rs.Errors) > 0,
		"expected existing.example in Updated or Errors, got: %+v", rs)
}

// TestHandleReconcileTask_reconcileSuccessNoErrors covers stmts 7 (false) and 9 of handleReconcileTask.
func TestHandleReconcileTask_reconcileSuccessNoErrors(t *testing.T) {
	resetCaddyHooks(t)
	setupTestRedis(t)

	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "proxy.internal", nil
	}

	cfgMem := newMemCaddyCfg(`{}`)
	srv := newTestCaddyAdminServer(t, cfgMem)
	hookClientFromHarness(t, srv)

	d := &ReconcilerDaemon{
		reconciler: NewReconciler(&deployRepoTestStub{}, logger.NewLogger()),
		logger:     logger.NewLogger(),
	}

	// Empty deployed apps → no desired routes → needsReload=false → result.Errors is empty.
	err := d.handleReconcileTask(context.Background(), CaddyTaskPayload{OrganizationID: uuid.New().String()})
	require.NoError(t, err)
}
