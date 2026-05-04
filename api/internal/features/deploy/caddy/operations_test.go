package caddy

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatDial(t *testing.T) {
	assert.Equal(t, "10.0.0.5:8443", FormatDial("10.0.0.5", 8443))
}

func Test_parseDial(t *testing.T) {
	host, port, err := parseDial("127.0.0.1:2019")
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1", host)
	assert.Equal(t, 2019, port)

	host, port, err = parseDial("[::1]:443")
	require.NoError(t, err)
	assert.Equal(t, "::1", host)
	assert.Equal(t, 443, port)

	_, _, err = parseDial("nohostport")
	require.Error(t, err)

	_, _, err = parseDial("h:badport")
	require.Error(t, err)

	_, _, err = parseDial("h:0")
	require.Error(t, err)

	_, _, err = parseDial("h:99999")
	require.Error(t, err)
}

func Test_resolveLogger(t *testing.T) {
	l := logger.NewLogger()
	out := resolveLogger(&l)
	assert.Equal(t, l, out)

	nilFallback := resolveLogger(nil)
	assert.NotNil(t, nilFallback)
}

func Test_jsonBody_Read(t *testing.T) {
	data := []byte(`{"hello":"world"}`)
	r := jsonReader(data)
	buf := make([]byte, 3)
	n, err := r.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	rest := make([]byte, 64)
	total := n
	for {
		n, err := r.Read(rest)
		total += n
		if err != nil {
			assert.True(t, errors.Is(err, io.EOF))
			break
		}
	}
	assert.Equal(t, len(data), total)

	buf2 := make([]byte, 16)
	n2, err2 := jsonReader(nil).Read(buf2)
	assert.Equal(t, 0, n2)
	assert.True(t, errors.Is(err2, io.EOF))
}

func Test_extractDomainFromRoute_missingHost(t *testing.T) {
	assert.Empty(t, extractDomainFromRoute(caddyhttp.Route{}))

	routeJSON := []byte(`{"match":[{"not_host":["x.example"]}],"handle":[]}`)
	var r caddyhttp.Route
	require.NoError(t, json.Unmarshal(routeJSON, &r))
	assert.Empty(t, extractDomainFromRoute(r))
}

func Test_extractUpstreamsFromRoute_skipsNonReverseProxy(t *testing.T) {
	hRaw, err := json.Marshal(map[string]string{"handler": "static_response"})
	require.NoError(t, err)
	r := caddyhttp.Route{HandlersRaw: []json.RawMessage{hRaw}}
	assert.Empty(t, extractUpstreamsFromRoute(r))

	badUpstream, err := json.Marshal(map[string]string{"handler": "reverse_proxy"})
	require.NoError(t, err)
	r2 := caddyhttp.Route{HandlersRaw: []json.RawMessage{badUpstream}}
	assert.Empty(t, extractUpstreamsFromRoute(r2))
}

func Test_extractDomainRoutes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		routes, err := extractDomainRoutes(&caddy.Config{})
		require.NoError(t, err)
		assert.Nil(t, routes)
	})

	t.Run("nixopus_server_routes", func(t *testing.T) {
		cfgJSON := []byte(`{
			"apps": {
				"http": {
					"servers": {
						"nixopus": {
							"routes": [
								{
									"match": [{"host": ["a.example.com"]}],
									"handle": [{"handler": "reverse_proxy", "upstreams": [{"dial": "10.1.1.1:8080"}]}]
								},
								{
									"match": [{"host": ["b.example.com"]}],
									"handle": [{"handler": "reverse_proxy", "upstreams": [{"dial": "10.1.1.2:8080"},{"dial": "10.1.1.3:9080"}]}]
								},
								{"match": [], "handle": []}
							]
						}
					}
				}
			}
		}`)
		var cfg caddy.Config
		require.NoError(t, json.Unmarshal(cfgJSON, &cfg))

		routes, err := extractDomainRoutes(&cfg)
		require.NoError(t, err)
		require.Len(t, routes, 2)

		var founda, foundb bool
		for _, rt := range routes {
			switch rt.Domain {
			case "a.example.com":
				founda = true
				assert.Equal(t, "10.1.1.1:8080", rt.UpstreamDial)
				assert.Nil(t, rt.Upstreams)
			case "b.example.com":
				foundb = true
				assert.Equal(t, []string{"10.1.1.2:8080", "10.1.1.3:9080"}, rt.Upstreams)
			}
		}
		assert.True(t, founda && foundb, "both domains must be extracted")
	})

	t.Run("invalid_http_app", func(t *testing.T) {
		cfg := &caddy.Config{
			AppsRaw: map[string]json.RawMessage{"http": json.RawMessage(`{`)},
		}
		_, err := extractDomainRoutes(cfg)
		require.Error(t, err)
	})
}

func Test_AddRemoveDomains_atomic_emptySlices(t *testing.T) {
	ctx := t.Context()
	assert.NoError(t, AddDomainsWithRetry(ctx, nil, nil, nil))
	assert.NoError(t, RemoveDomainsWithRetry(ctx, nil, nil, nil))
	assert.NoError(t, AddDomainsAtomic(ctx, nil, nil, nil))
	assert.NoError(t, RemoveDomainsAtomic(ctx, nil, nil, nil))
}

func Test_extDomainsKey_and_pendingRemovalKey_format(t *testing.T) {
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	assert.Equal(t, "caddy:ext_domains:"+id.String(), extDomainsKey(id))
	assert.Equal(t, "caddy:pending_removals:"+id.String(), pendingRemovalKey(id))
}
