package auth

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
)

func saveBetterAuth(cfg types.BetterAuthConfig) func() {
	prev := config.AppConfig.BetterAuth
	config.AppConfig.BetterAuth = cfg
	return func() { config.AppConfig.BetterAuth = prev }
}

func saveHTTPClient(c *http.Client) func() {
	prev := HTTPClient
	HTTPClient = c
	return func() { HTTPClient = prev }
}

type errRoundTripper struct{ err error }

func (e errRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, e.err
}

func Test_getBetterAuthURL(t *testing.T) {
	t.Run("empty falls back localhost", func(t *testing.T) {
		revert := saveBetterAuth(types.BetterAuthConfig{URL: "", Secret: "s"})
		t.Cleanup(revert)
		require.Equal(t, "http://localhost:9090", getBetterAuthURL())
	})
	t.Run("bare host HTTPS prefix", func(t *testing.T) {
		revert := saveBetterAuth(types.BetterAuthConfig{URL: "auth.example.com", Secret: "s"})
		t.Cleanup(revert)
		require.Equal(t, "https://auth.example.com", getBetterAuthURL())
	})
	t.Run("preserve http scheme", func(t *testing.T) {
		revert := saveBetterAuth(types.BetterAuthConfig{URL: "http://localhost:9090/", Secret: "s"})
		t.Cleanup(revert)
		require.Equal(t, "http://localhost:9090/", getBetterAuthURL())
	})
}

func Test_getBetterAuthAPI(t *testing.T) {
	revert := saveBetterAuth(types.BetterAuthConfig{URL: "http://localhost:9191", Secret: "s"})
	t.Cleanup(revert)
	require.Equal(t, "http://localhost:9191/api/auth", getBetterAuthAPI())
}

func Test_extractOriginFromReferer(t *testing.T) {
	require.Equal(t, "", extractOriginFromReferer("/relative"))
	require.Equal(t, "", extractOriginFromReferer("ftp://bad"))
	require.Equal(t, "https://app.example.com", extractOriginFromReferer("https://app.example.com/path?x=1"))
	require.Equal(t, "http://local.test", extractOriginFromReferer("http://local.test"))
}

func Test_forwardCookies_cookieHeader(t *testing.T) {
	orig := httptest.NewRequest(http.MethodGet, "/", nil)
	orig.Header.Set("Cookie", "a=1; b=2")
	target := httptest.NewRequest(http.MethodGet, "/", nil)
	forwardCookies(orig, target)
	require.Equal(t, "a=1; b=2", target.Header.Get("Cookie"))
}

func Test_forwardCookies_individualCookies(t *testing.T) {
	orig := httptest.NewRequest(http.MethodGet, "/", nil)
	orig.AddCookie(&http.Cookie{Name: "n", Value: "v"})
	target := httptest.NewRequest(http.MethodGet, "/", nil)
	forwardCookies(orig, target)
	cookies := target.Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "n", cookies[0].Name)
	require.Equal(t, "v", cookies[0].Value)
}

func Test_forwardCookies_listOverrideEmptyHeader(t *testing.T) {
	prev := forwardCookiesList
	t.Cleanup(func() { forwardCookiesList = prev })
	forwardCookiesList = func(_ *http.Request) []*http.Cookie {
		return []*http.Cookie{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}}
	}

	orig := httptest.NewRequest(http.MethodGet, "/", nil)
	target := httptest.NewRequest(http.MethodGet, "/", nil)
	forwardCookies(orig, target)
	require.Len(t, target.Cookies(), 2)
}

func Test_forwardCookies_warnNoCookies(t *testing.T) {
	orig := httptest.NewRequest(http.MethodGet, "/", nil)
	target := httptest.NewRequest(http.MethodGet, "/", nil)
	forwardCookies(orig, target)
	require.Empty(t, target.Cookies())
}

func Test_forwardHeaders(t *testing.T) {
	orig := httptest.NewRequest(http.MethodGet, "https://api.example/foo", nil)
	orig.Header.Set("Authorization", "Bearer t")
	orig.Header.Set("Origin", "https://orig.example")
	orig.Header.Set("x-api-key", "k")
	orig.Header.Set("X-Forwarded-Proto", "https")
	orig.Header.Set("X-Forwarded-Host", "api.example")

	target := httptest.NewRequest(http.MethodGet, "/", nil)
	forwardHeaders(orig, target)

	require.Equal(t, "Bearer t", target.Header.Get("Authorization"))
	require.Equal(t, "https://orig.example", target.Header.Get("Origin"))
	require.Equal(t, "k", target.Header.Get("x-api-key"))
	require.Equal(t, "https", target.Header.Get("X-Forwarded-Proto"))
	require.Equal(t, "api.example", target.Header.Get("X-Forwarded-Host"))
	require.Equal(t, "application/json", target.Header.Get("Content-Type"))
}

func Test_forwardHeaders_refererDerivesOrigin(t *testing.T) {
	orig := httptest.NewRequest(http.MethodGet, "/", nil)
	orig.Header.Set("Referer", "https://app.example/dashboard")
	target := httptest.NewRequest(http.MethodGet, "/", nil)
	forwardHeaders(orig, target)
	require.Equal(t, "https://app.example", target.Header.Get("Origin"))
	require.Equal(t, "https://app.example/dashboard", target.Header.Get("Referer"))
}

func Test_forwardHeaders_httpsFromTLS(t *testing.T) {
	orig := httptest.NewRequest(http.MethodGet, "/", nil)
	orig.TLS = &tls.ConnectionState{}
	target := httptest.NewRequest(http.MethodGet, "/", nil)
	forwardHeaders(orig, target)
	require.Equal(t, "https", target.Header.Get("X-Forwarded-Proto"))
}

func Test_forwardHeaders_httpsFromOriginHeader(t *testing.T) {
	orig := httptest.NewRequest(http.MethodGet, "/", nil)
	orig.Header.Set("Origin", "https://x.example")
	target := httptest.NewRequest(http.MethodGet, "/", nil)
	forwardHeaders(orig, target)
	require.Equal(t, "https", target.Header.Get("X-Forwarded-Proto"))
}

func Test_parseSessionResponse(t *testing.T) {
	makeReq := func() (*http.Request, *http.Request) {
		or := httptest.NewRequest(http.MethodGet, "/", nil)
		or.AddCookie(&http.Cookie{Name: "sid", Value: "1"})
		tg := httptest.NewRequest(http.MethodGet, "/", nil)
		return tg, or
	}

	t.Run("null body", func(t *testing.T) {
		target, original := makeReq()
		_, err := parseSessionResponse([]byte("null"), 200, "http://u", target, original)
		require.Error(t, err)
		require.Contains(t, err.Error(), "null")
	})

	t.Run("empty body", func(t *testing.T) {
		target, original := makeReq()
		target.Header.Set("Cookie", "a=b")
		_, err := parseSessionResponse([]byte("   "), 401, "http://u", target, original)
		require.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		target, original := makeReq()
		_, err := parseSessionResponse([]byte(`{`), 200, "http://u", target, original)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse")
	})

	t.Run("error field set", func(t *testing.T) {
		target, original := makeReq()
		body := []byte(`{"error":{"message":"nope","status":401},"user":{"id":"","email":""}}`)
		_, err := parseSessionResponse(body, 200, "http://u", target, original)
		require.Error(t, err)
		require.Contains(t, err.Error(), "nope")
	})

	t.Run("missing user id", func(t *testing.T) {
		target, original := makeReq()
		body := []byte(`{"session":{"id":"s"},"user":{"id":"","email":"e@x"}}`)
		_, err := parseSessionResponse(body, 200, "http://u", target, original)
		require.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		target, original := makeReq()
		body := []byte(`{"session":{"id":"s","userId":"u"},"user":{"id":"u","email":"e@x"}}`)
		sr, err := parseSessionResponse(body, 200, "http://u", target, original)
		require.NoError(t, err)
		require.Equal(t, "u", sr.User.ID)
	})
}

func TestVerifySession_newRequestFailure(t *testing.T) {
	revert := saveBetterAuth(types.BetterAuthConfig{
		URL:    "http://[\n",
		Secret: "s",
	})
	t.Cleanup(revert)

	_, err := VerifySession(httptest.NewRequest(http.MethodGet, "/", nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), "create request")
}

func TestVerifySession_httpDoFails(t *testing.T) {
	revertBa := saveBetterAuth(types.BetterAuthConfig{
		URL:    "http://127.0.0.1:1",
		Secret: "s",
	})
	t.Cleanup(revertBa)
	revertHc := saveHTTPClient(&http.Client{Transport: errRoundTripper{err: errors.New("down")}})
	t.Cleanup(revertHc)

	_, err := VerifySession(httptest.NewRequest(http.MethodGet, "/", nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), "verify session")
}

func TestVerifySession_readBodyFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	t.Cleanup(srv.Close)

	revertBa := saveBetterAuth(types.BetterAuthConfig{URL: srv.URL, Secret: "s"})
	t.Cleanup(revertBa)
	prevRead := authReadAll
	authReadAll = func(r io.Reader) ([]byte, error) {
		return nil, errors.New("read fail")
	}
	t.Cleanup(func() { authReadAll = prevRead })

	_, err := VerifySession(httptest.NewRequest(http.MethodGet, "/", nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), "read")
}

func sessionJSON(uid, activeOrg string) []byte {
	m := map[string]any{
		"session": map[string]any{
			"id": "sid", "userId": uid, "expiresAt": "", "token": "t",
		},
		"user": map[string]any{
			"id": uid, "email": "e@x", "name": "n", "emailVerified": true,
		},
	}
	if activeOrg != "" {
		m["session"].(map[string]any)["activeOrganizationId"] = activeOrg
	}
	b, _ := json.Marshal(m)
	return b
}

func TestVerifySession_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/auth/get-session", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(sessionJSON("user-1", ""))
	}))
	t.Cleanup(srv.Close)

	revertBa := saveBetterAuth(types.BetterAuthConfig{URL: srv.URL, Secret: "s"})
	t.Cleanup(revertBa)
	revertHc := saveHTTPClient(srv.Client())
	t.Cleanup(revertHc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example")
	resp, err := VerifySession(req)
	require.NoError(t, err)
	require.Equal(t, "user-1", resp.User.ID)
}

func Test_forwardHeaders_userAgent(t *testing.T) {
	orig := httptest.NewRequest(http.MethodGet, "/", nil)
	orig.Header.Set("User-Agent", customUA)
	target := httptest.NewRequest(http.MethodGet, "/", nil)
	forwardHeaders(orig, target)
	require.Equal(t, customUA, target.Header.Get("User-Agent"))
}

const customUA = "unit-test-agent/1"
