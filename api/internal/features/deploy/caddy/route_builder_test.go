package caddy

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func Test_mapRoutingStrategy(t *testing.T) {
	assert.Equal(t, "round_robin", mapRoutingStrategy(shared_types.RoutingStrategyRoundRobin))
	assert.Equal(t, "first", mapRoutingStrategy(shared_types.RoutingStrategyPrimaryFailover))
	assert.Equal(t, "", mapRoutingStrategy(shared_types.RoutingStrategySingle))
	assert.Equal(t, "", mapRoutingStrategy(shared_types.RoutingStrategyPerServerDomain))
}

func Test_orderPrimaryFirst(t *testing.T) {
	u := []string{"a", "b", "c"}
	assert.Equal(t, []string{"a", "b", "c"}, orderPrimaryFirst(u, 0))
	assert.Equal(t, []string{"a", "b", "c"}, orderPrimaryFirst(u, -1))
	assert.Equal(t, []string{"a", "b", "c"}, orderPrimaryFirst(u, 3))

	assert.Equal(t, []string{"b", "a", "c"}, orderPrimaryFirst([]string{"a", "b", "c"}, 1))
}

func Test_buildSingleUpstreamRoutes(t *testing.T) {
	up := []shared_types.ApplicationDomain{{Domain: "x.example"}, {Domain: ""}}
	routes := buildSingleUpstreamRoutes(up, "proxy.internal", 9000, shared_types.DockerFile)
	require.Len(t, routes, 1)
	assert.Equal(t, "proxy.internal:9000", routes[0].UpstreamDial)

	p90 := 90
	svcPort := shared_types.ComposeService{Port: 4000}
	domCompose := []shared_types.ApplicationDomain{
		{Domain: "c.example", ComposeService: &svcPort},
		{Domain: "orphan.compose", ComposeService: nil},
		{
			Domain:         "override.example",
			ComposeService: nil,
			Port:           &p90,
		},
	}
	rc := buildSingleUpstreamRoutes(domCompose, "ignore", 0, shared_types.DockerCompose)
	require.Len(t, rc, 2)
	for _, rr := range rc {
		if rr.Domain == "c.example" {
			assert.Equal(t, "ignore:4000", rr.UpstreamDial)
		}
		if rr.Domain == "override.example" {
			assert.Equal(t, "ignore:90", rr.UpstreamDial)
		}
	}
}

func Test_resolveServerHosts_defaultVsMultiProxy(t *testing.T) {
	l := logger.NewLogger()
	h1 := "internal-1.proxy"
	h2 := "internal-2.proxy"
	pub := "pub.example.net"

	def := shared_types.ApplicationServer{
		IsPrimary: true,
		Server: &shared_types.SSHKey{
			Host:      strPtr("raw-a"),
			ProxyHost: strPtr(pub),
			IsDefault: true,
		},
	}

	servers := []shared_types.ApplicationServer{
		def,
		{
			Server: &shared_types.SSHKey{
				Host:      strPtr(h1),
				IsDefault: false,
			},
		},
		{
			IsPrimary: false,
			Server: &shared_types.SSHKey{
				Host:      strPtr(h2),
				IsDefault: false,
			},
		},
	}

	hosts, primaryIdx := resolveServerHosts([]shared_types.ApplicationServer{{Server: nil}}, &l)
	assert.Empty(t, hosts)
	assert.Equal(t, -1, primaryIdx)

	hosts, primaryIdx = resolveServerHosts(servers, &l)
	assert.Equal(t, 0, primaryIdx)
	require.Len(t, hosts, 3)
	assert.Equal(t, pub, hosts[0])
	assert.Equal(t, h1, hosts[1])
	assert.Equal(t, h2, hosts[2])

	multiHosts, _ := resolveServerHosts(servers[1:], &l)
	require.Len(t, multiHosts, 2)
	assert.Equal(t, h1, multiHosts[0])
	assert.Equal(t, h2, multiHosts[1])
}

func TestBuildMultiUpstreamRoutes_singleServer_fallback(t *testing.T) {
	l := logger.NewLogger()
	app := shared_types.Application{
		ID: uuid.New(),
	}
	stub := &deployRepoTestStub{servers: []shared_types.ApplicationServer{{}}}
	doms := []shared_types.ApplicationDomain{{Domain: "app.example"}}

	got := BuildMultiUpstreamRoutes(context.Background(), stub, &l, app, doms, "10.10.10.10", 7000)
	require.Len(t, got, 1)
	assert.Equal(t, "app.example", got[0].Domain)
	assert.Equal(t, "10.10.10.10:7000", got[0].UpstreamDial)
	assert.Empty(t, got[0].Upstreams)
}

func TestBuildMultiUpstreamRoutes_GetApplicationServersErr_falls_back(t *testing.T) {
	l := logger.NewLogger()
	app := shared_types.Application{ID: uuid.New()}
	stub := &deployRepoTestStub{err: assert.AnError}
	doms := []shared_types.ApplicationDomain{{Domain: "e.example"}}

	got := BuildMultiUpstreamRoutes(context.Background(), stub, &l, app, doms, "fallback.host", 5000)
	require.Len(t, got, 1)
	assert.Equal(t, "fallback.host:5000", got[0].UpstreamDial)
}

func TestBuildMultiUpstreamRoutes_multi_upstream_roundRobin(t *testing.T) {
	l := logger.NewLogger()
	app := shared_types.Application{
		ID:              uuid.New(),
		RoutingStrategy: shared_types.RoutingStrategyRoundRobin,
		BuildPack:       shared_types.DockerFile,
	}
	hA, hB := "10.0.0.1", "10.0.0.2"
	stub := &deployRepoTestStub{
		servers: []shared_types.ApplicationServer{
			{IsPrimary: true, Server: &shared_types.SSHKey{Host: strPtr(hA), IsDefault: false}},
			{Server: &shared_types.SSHKey{Host: strPtr(hB), IsDefault: false}},
		},
	}
	doms := []shared_types.ApplicationDomain{{Domain: "lb.example"}}

	got := BuildMultiUpstreamRoutes(context.Background(), stub, &l, app, doms, "unused", 3000)
	require.Len(t, got, 1)
	assert.Equal(t, "lb.example", got[0].Domain)
	assert.Equal(t, "round_robin", got[0].LBPolicy)
	require.Len(t, got[0].Upstreams, 2)
	for _, dial := range got[0].Upstreams {
		assert.True(t, strings.HasSuffix(dial, ":3000"))
	}
}

func TestBuildMultiUpstreamRoutes_multi_hostsEmpty_fallsBackToSingleUpstream(t *testing.T) {
	l := logger.NewLogger()
	app := shared_types.Application{
		ID:              uuid.New(),
		RoutingStrategy: shared_types.RoutingStrategyRoundRobin,
	}
	stub := &deployRepoTestStub{
		servers: []shared_types.ApplicationServer{
			{Server: &shared_types.SSHKey{}},
			{Server: nil},
			{},
		},
	}
	got := BuildMultiUpstreamRoutes(context.Background(), stub, &l, app, []shared_types.ApplicationDomain{{Domain: "skip.me"}}, "host", 1)
	require.Len(t, got, 1)
	assert.Equal(t, "skip.me", got[0].Domain)
	assert.Equal(t, "host:1", got[0].UpstreamDial)
}
