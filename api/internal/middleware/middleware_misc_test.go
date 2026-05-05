package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddleware_panicAndOK(t *testing.T) {
	h := RecoveryMiddleware(logger.NewLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panic" {
			panic("boom")
		}
		w.WriteHeader(http.StatusTeapot)
	}))

	t.Run("panic", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "Internal server error")
	})
	t.Run("ok", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ok", nil))
		require.Equal(t, http.StatusTeapot, rec.Code)
	})
}

func TestServerIDMiddleware(t *testing.T) {
	h := ServerIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(types.ServerIDKey)
		_, _ = fmt.Fprintf(w, "%v", v)
	}))

	t.Run("invalid uuid ignored", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/?server_id=not-uuid", nil)
		h.ServeHTTP(rec, req)
		require.Equal(t, "<nil>", rec.Body.String())
	})
	t.Run("query uuid", func(t *testing.T) {
		id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/?server_id="+id, nil)
		h.ServeHTTP(rec, req)
		require.Equal(t, id, rec.Body.String())
	})
	t.Run("header uuid", func(t *testing.T) {
		id := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Server-Id", id)
		h.ServeHTTP(rec, req)
		require.Equal(t, id, rec.Body.String())
	})
}

func TestCorsMiddleware(t *testing.T) {
	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })
	config.AppConfig = types.Config{
		CORS: types.CORSConfig{AllowedOrigin: "https://app.example.com, https://other.example.com "},
	}

	h := CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	t.Run("webhook bypass", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhook/", nil)
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("options", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/x", nil)
		req.Header.Set("Origin", "https://app.example.com")
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("websocket upgrade", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Origin", "http://localhost:3000")
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("allowed origin from config", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/r", nil)
		req.Header.Set("Origin", "https://other.example.com")
		h.ServeHTTP(rec, req)
		require.Equal(t, "https://other.example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("localhost dev origin", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/r", nil)
		req.Header.Set("Origin", "http://localhost:7443")
		h.ServeHTTP(rec, req)
		require.Equal(t, "http://localhost:7443", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("disallowed origin omits allow-origin", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/r", nil)
		req.Header.Set("Origin", "https://evil.example")
		h.ServeHTTP(rec, req)
		require.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("empty cors config still allows localhost append", func(t *testing.T) {
		config.AppConfig = types.Config{CORS: types.CORSConfig{AllowedOrigin: ""}}
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/r", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		h.ServeHTTP(rec, req)
		require.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

type hijackRecorder struct {
	*httptest.ResponseRecorder
}

func (h *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func TestLoggingMiddleware_andHelpers(t *testing.T) {
	t.Run("websocket branch", func(t *testing.T) {
		h := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusSwitchingProtocols)
		}))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		req.RemoteAddr = "127.0.0.1:1"
		h.ServeHTTP(rec, req)
	})

	t.Run("http log branches", func(t *testing.T) {
		h := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(1100 * time.Millisecond)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("hello"))
		}))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/slow", nil)
		h.ServeHTTP(rec, req)
	})

	t.Run("status and method colors", func(t *testing.T) {
		for _, tc := range []struct {
			code   int
			method string
		}{
			{100, "TRACE"},
			{201, http.MethodGet},
			{301, http.MethodPost},
			{401, http.MethodPut},
			{501, http.MethodDelete},
			{200, http.MethodPatch},
		} {
			_ = getStatusColor(tc.code)
			_ = getMethodColor(tc.method)
		}
		require.Equal(t, "GET    ", padMethod(http.MethodGet))
	})

	t.Run("responseWriter hijack not supported", func(t *testing.T) {
		base := httptest.NewRecorder()
		rw := newResponseWriter(base)
		_, _, err := rw.Hijack()
		require.Equal(t, http.ErrNotSupported, err)
	})

	t.Run("responseWriter hijack supported", func(t *testing.T) {
		hr := &hijackRecorder{ResponseRecorder: httptest.NewRecorder()}
		rw := newResponseWriter(hr)
		_, _, err := rw.Hijack()
		require.NoError(t, err)
	})

	t.Run("responseWriter write counts bytes", func(t *testing.T) {
		base := httptest.NewRecorder()
		rw := newResponseWriter(base)
		rw.WriteHeader(http.StatusBadRequest)
		n, err := rw.Write([]byte("ab"))
		require.NoError(t, err)
		require.Equal(t, 2, n)
		require.Equal(t, 400, rw.statusCode)
		require.Equal(t, int64(2), rw.length)
	})

	t.Run("microsecond duration branch", func(t *testing.T) {
		rw := newResponseWriter(httptest.NewRecorder())
		logHTTPRequest(rw, httptest.NewRequest(http.MethodGet, "/", nil), time.Now().Add(-500*time.Microsecond))
	})
}

func Test_extractIP(t *testing.T) {
	require.Equal(t, "1.2.3.4", extractIP("1.2.3.4:5678"))
	require.Contains(t, extractIP("[::1]:1234"), "[::1]")
}

func Test_cleanupClients_removesStale(t *testing.T) {
	clientsMtx.Lock()
	clients["1.1.1.1"] = &client{
		limiter:  nil,
		lastSeen: time.Now().Add(-10 * time.Minute),
	}
	clientsMtx.Unlock()
	cleanupClients()
	clientsMtx.Lock()
	_, ok := clients["1.1.1.1"]
	clientsMtx.Unlock()
	require.False(t, ok)
	clearRateLimiterClients(t)
}

func TestRateLimiter_exhausted(t *testing.T) {
	clearRateLimiterClients(t)
	h := RateLimiter(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	const ip = "9.9.9.9:9999"
	for i := 0; i < 200; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			require.Equal(t, http.StatusTooManyRequests, rec.Code)
			clearRateLimiterClients(t)
			return
		}
	}
	require.Fail(t, "expected 429")
}

func TestNewRateLimiterWithConfig_exhausted(t *testing.T) {
	h := NewRateLimiterWithConfig(0.001, 1)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	ip := "8.8.8.8:8888"
	for i := 0; i < 50; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			clearRateLimiterClients(t)
			return
		}
	}
	clearRateLimiterClients(t)
	require.Fail(t, "expected 429 from config limiter")
}

func Test_isWebSocketUpgrade(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	require.False(t, isWebSocketUpgrade(r))
	r.Header.Set("Upgrade", "websocket")
	require.False(t, isWebSocketUpgrade(r))
	r.Header.Set("Connection", "Upgrade")
	require.True(t, isWebSocketUpgrade(r))
}

func Test_rbacPureHelpers(t *testing.T) {
	require.Equal(t, "res:read", buildRequiredPermission("res", http.MethodGet))
	require.Equal(t, "res:create", buildRequiredPermission("res", http.MethodPost))
	require.Equal(t, "res:update", buildRequiredPermission("res", http.MethodPut))
	require.Equal(t, "res:update", buildRequiredPermission("res", http.MethodPatch))
	require.Equal(t, "res:delete", buildRequiredPermission("res", http.MethodDelete))
	require.Equal(t, "res:read", buildRequiredPermission("res", "OTHER"))

	require.True(t, hasPermission([]string{"a", "b"}, "b"))
	require.False(t, hasPermission([]string{"a"}, "b"))

	pref := buildOrgRolePrefix("org-1")
	require.Equal(t, "orgid_org-1_", pref)
	roles := filterOrganizationRolesFromStrings([]string{pref + "custom", "owner", "nope"}, "org-1")
	require.Contains(t, roles, "owner")
	require.Contains(t, roles, pref+"custom")

	perms := getPermissionsForRole("unknown-role-xyz")
	require.Equal(t, rolePermissions["viewer"], perms)

	InitRBACCache(nil)
	cachePermissions("u", "o", []string{"member"}, []string{"machine:read"})
}
