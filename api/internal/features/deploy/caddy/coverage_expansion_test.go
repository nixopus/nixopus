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
