package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/nixopus/nixopus/api/internal/cache"
	betterauth "github.com/nixopus/nixopus/api/internal/features/auth"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
)

func testRedisCache(t *testing.T) *cache.Cache {
	t.Helper()
	mr := miniredis.RunT(t)
	c, err := cache.NewCache("redis://" + mr.Addr())
	require.NoError(t, err)
	return c
}

func authReq(t *testing.T, path string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.AddCookie(&http.Cookie{Name: "mw_test_sid", Value: t.Name()})
	return r
}

func sessionResp(uid, orgID string) *betterauth.SessionResponse {
	var r betterauth.SessionResponse
	r.User.ID = uid
	if orgID != "" {
		r.Session.ActiveOrganizationID = &orgID
	}
	return &r
}

func Test_sessionCacheKey_stable(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Cookie", "a=b")
	r.Header.Set("Authorization", "Bearer x")
	r.Header.Set("x-api-key", "k")
	k1 := sessionCacheKey(r)
	k2 := sessionCacheKey(r)
	require.Equal(t, k1, k2)
}

func Test_verifySessionWithFallback_andCached(t *testing.T) {
	orig := verifySessionFn
	t.Cleanup(func() { verifySessionFn = orig })

	t.Run("api key fallback", func(t *testing.T) {
		var verifyCalls atomic.Int32
		verifySessionFn = func(r *http.Request) (*betterauth.SessionResponse, error) {
			// First call is the original request; second is the synthetic API-key probe built by verifySessionWithFallback.
			if verifyCalls.Add(1) == 2 && r.Header.Get("x-api-key") == "secret" {
				return sessionResp("u1", "o1"), nil
			}
			return nil, errors.New("no cookie session")
		}
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("x-api-key", "secret")
		r.Header.Set("Origin", "https://app")
		r.Header.Set("X-Forwarded-Proto", "https")
		got, err := verifySessionWithFallback(r)
		require.NoError(t, err)
		require.Equal(t, "u1", got.User.ID)
	})

	t.Run("new request for api key error", func(t *testing.T) {
		prev := authNewRequestForAPIKey
		authNewRequestForAPIKey = func(method, url string, body io.Reader) (*http.Request, error) {
			return nil, errors.New("bad newrequest")
		}
		t.Cleanup(func() { authNewRequestForAPIKey = prev })
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return nil, errors.New("fail")
		}
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("x-api-key", "k")
		_, err := verifySessionWithFallback(r)
		require.ErrorContains(t, err, "failed to create API key request")
	})
}

func Test_verifySessionCached_paths(t *testing.T) {
	origVerify := verifySessionFn
	origMarshal := jsonMarshalSessionResponse
	t.Cleanup(func() {
		verifySessionFn = origVerify
		jsonMarshalSessionResponse = origMarshal
	})

	c := testRedisCache(t)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(ctx)

	t.Run("nil cache uses fallback only", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp("a", "b"), nil
		}
		got, err := verifySessionCached(r, nil)
		require.NoError(t, err)
		require.Equal(t, "a", got.User.ID)
	})

	t.Run("cache hit", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return nil, errors.New("should not call")
		}
		key := sessionCacheKey(r)
		b, err := json.Marshal(sessionResp("cached-user", "org"))
		require.NoError(t, err)
		require.NoError(t, c.SetSession(ctx, key, b))
		got, err := verifySessionCached(r, c)
		require.NoError(t, err)
		require.Equal(t, "cached-user", got.User.ID)
	})

	t.Run("bad json in cache falls back", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp("fresh", "org"), nil
		}
		r2 := httptest.NewRequest(http.MethodGet, "/", nil)
		r2 = r2.WithContext(ctx)
		key := sessionCacheKey(r2)
		require.NoError(t, c.SetSession(ctx, key, []byte("{")))
		_, err := verifySessionCached(r2, c)
		require.NoError(t, err)
	})

	t.Run("marshal failure skips set session", func(t *testing.T) {
		jsonMarshalSessionResponse = func(any) ([]byte, error) {
			return nil, errors.New("no marshal")
		}
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp("u", "00000000-0000-0000-0000-000000000099"), nil
		}
		r3 := httptest.NewRequest(http.MethodGet, "/", nil)
		r3 = r3.WithContext(ctx)
		_, err := verifySessionCached(r3, c)
		require.NoError(t, err)
		jsonMarshalSessionResponse = origMarshal
	})
}

func Test_isAuthEndpoint_and_extractOrgIDFromSession(t *testing.T) {
	require.True(t, isAuthEndpoint("/api/v1/auth/login"))
	require.True(t, isAuthEndpoint("/api/auth/foo"))
	require.True(t, isAuthEndpoint("/api/v1/user/organizations/extra"))
	require.False(t, isAuthEndpoint("/api/v1/machines"))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Organization-Id", "hdr-org")
	require.Equal(t, "hdr-org", extractOrgIDFromSession(nil, r))

	s := sessionResp("u", "sess-org")
	require.Equal(t, "sess-org", extractOrgIDFromSession(s, r))
}

func Test_resolveAndVerifyOrganization_DB_fallback(t *testing.T) {
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	uid := uuid.New()
	oid := uuid.New()
	insertUser(t, db, ctx, uid, "m@x")
	insertMember(t, db, ctx, uid, oid)

	got, err := resolveAndVerifyOrganization(ctx, httptest.NewRequest(http.MethodGet, "/", nil), nil, app, uid.String(), sessionResp(uid.String(), ""))
	require.NoError(t, err)
	require.Equal(t, oid.String(), got)
}

func Test_resolveAndVerifyOrganization_errors(t *testing.T) {
	app := testApp(t)
	ctx := app.Ctx
	t.Setenv("AUTH_SERVICE_URL", "http://127.0.0.1:9")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })
	_, err := resolveAndVerifyOrganization(ctx, httptest.NewRequest(http.MethodGet, "/", nil), nil, app, uuid.New().String(), sessionResp(uuid.New().String(), "00000000-0000-0000-0000-000000000001"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong")
}

func Test_verifyOrganizationMembership_cacheAndAPI(t *testing.T) {
	c := testRedisCache(t)
	ctx := context.Background()
	uid := uuid.New().String()
	oid := uuid.New().String()

	require.NoError(t, c.SetOrgMembership(ctx, uid, oid, true))
	ok, err := verifyOrganizationMembership(ctx, httptest.NewRequest(http.MethodGet, "/", nil), c, uid, oid)
	require.NoError(t, err)
	require.True(t, ok)

	orig := betterauth.HTTPClient
	betterauth.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("nope")
		}),
	}
	t.Cleanup(func() { betterauth.HTTPClient = orig })

	ok, err = verifyOrganizationMembership(ctx, httptest.NewRequest(http.MethodGet, "/", nil), c, uid, "00000000-0000-0000-0000-000000000002")
	require.Error(t, err)
	require.False(t, ok)
}

func Test_AuthMiddleware_branches(t *testing.T) {
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	uid := uuid.New()
	oid := uuid.New()
	insertUser(t, db, ctx, uid, "auth@x")
	insertMember(t, db, ctx, uid, oid)

	c := testRedisCache(t)
	InitRBACCache(c)
	t.Cleanup(func() { InitRBACCache(nil) })

	orig := verifySessionFn
	t.Cleanup(func() { verifySessionFn = orig })

	t.Run("unauthorized session", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return nil, errors.New("nope")
		}
		h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, c)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, authReq(t, "/api/v1/machines"))
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid better auth user id", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp("not-a-uuid", oid.String()), nil
		}
		h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, c)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, authReq(t, "/api/v1/machines"))
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("user not in database", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp(uuid.New().String(), oid.String()), nil
		}
		h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, c)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, authReq(t, "/api/v1/machines"))
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("success with org resolution", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp(uid.String(), oid.String()), nil
		}
		require.NoError(t, c.SetOrgMembership(ctx, uid.String(), oid.String(), true))
		var saw bool
		h := AuthMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			v, _ := r.Context().Value(types.OrganizationIDKey).(string)
			saw = v == oid.String()
		}), app, c)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, authReq(t, "/api/v1/machines"))
		require.True(t, saw, "body=%s code=%d", rec.Body.String(), rec.Code)
	})

	t.Run("auth endpoint skips org", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp(uid.String(), ""), nil
		}
		var org any
		h := AuthMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			org = r.Context().Value(types.OrganizationIDKey)
		}), app, c)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, authReq(t, "/api/v1/auth/login"))
		require.Nil(t, org)
		require.Equal(t, http.StatusOK, rec.Code)
	})
}

func Test_tryM2MJWTAuth(t *testing.T) {
	resetJWKSState(t)
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	uid := uuid.New()
	oid := uuid.New()
	insertUser(t, db, ctx, uid, "m2m@x")

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privJWK, err := jwk.FromRaw(key)
	require.NoError(t, err)
	require.NoError(t, privJWK.Set(jwk.KeyIDKey, "kid1"))
	require.NoError(t, privJWK.Set(jwk.AlgorithmKey, jwa.RS256))
	pubJWK, err := jwk.PublicKeyOf(privJWK)
	require.NoError(t, err)
	pubSet := jwk.NewSet()
	require.NoError(t, pubSet.AddKey(pubJWK))

	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(pubSet)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_JWKS_URL", srv.URL+"/jwks")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_JWKS_URL") })

	now := time.Now()
	tokBuilt, err := jwt.NewBuilder().
		Subject("m2m").
		IssuedAt(now.Add(-time.Minute)).
		NotBefore(now.Add(-time.Minute)).
		Expiration(now.Add(time.Hour)).
		Claim("https://nixopus.com/org", oid.String()).
		Build()
	require.NoError(t, err)
	ph := jws.NewHeaders()
	require.NoError(t, ph.Set(jws.KeyIDKey, "kid1"))
	raw, err := jwt.Sign(tokBuilt, jwt.WithKey(jwa.RS256, key, jws.WithProtectedHeaders(ph)))
	require.NoError(t, err)
	tok := string(raw)

	t.Run("success", func(t *testing.T) {
		var gotOrg any
		h := AuthMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			gotOrg = r.Context().Value(types.OrganizationIDKey)
		}), app, nil)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/x", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("X-Organization-Id", oid.String())
		req.Header.Set("X-User-Id", uid.String())
		h.ServeHTTP(rec, req)
		require.Equal(t, oid.String(), gotOrg)
	})

	t.Run("missing X-User-Id", func(t *testing.T) {
		h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, nil)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/x", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("X-Organization-Id", oid.String())
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid X-User-Id", func(t *testing.T) {
		h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, nil)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/x", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("X-Organization-Id", oid.String())
		req.Header.Set("X-User-Id", "bad")
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("non jwt bearer falls through to session", func(t *testing.T) {
		orig := verifySessionFn
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp(uid.String(), oid.String()), nil
		}
		t.Cleanup(func() { verifySessionFn = orig })
		rc := testRedisCache(t)
		require.NoError(t, rc.SetOrgMembership(context.Background(), uid.String(), oid.String(), true))
		var ok bool
		h := AuthMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			ok = r.Context().Value(types.OrganizationIDKey) == oid.String()
		}), app, rc)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/machines", nil)
		req.Header.Set("Authorization", "Bearer opaque-not-jwt")
		h.ServeHTTP(rec, req)
		require.True(t, ok)
	})
}

func Test_RBACMiddleware(t *testing.T) {
	c := testRedisCache(t)
	InitRBACCache(c)
	t.Cleanup(func() { InitRBACCache(nil) })

	uid := uuid.New()
	oid := uuid.New()
	u := &types.User{ID: uid, Email: "rbac@x", Name: "n"}
	perms := append([]string{}, rolePermissions["owner"]...)
	require.NoError(t, c.SetRBACPermissions(context.Background(), uid.String(), oid.String(), &cache.CachedRBACPermissions{
		Roles:       []string{"owner"},
		Permissions: perms,
	}))

	t.Run("missing org header uses context", func(t *testing.T) {
		h := RBACMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &storage.App{}, "machine")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/machines", nil)
		ctx := context.WithValue(context.Background(), types.UserContextKey, u)
		ctx = context.WithValue(ctx, types.OrganizationIDKey, oid.String())
		h.ServeHTTP(rec, req.WithContext(ctx))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing user", func(t *testing.T) {
		h := RBACMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &storage.App{}, "machine")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Organization-Id", oid.String())
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid user type", func(t *testing.T) {
		h := RBACMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &storage.App{}, "machine")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Organization-Id", oid.String())
		ctx := context.WithValue(context.Background(), types.UserContextKey, "not-user")
		h.ServeHTTP(rec, req.WithContext(ctx))
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		require.NoError(t, c.SetRBACPermissions(context.Background(), uid.String(), oid.String(), &cache.CachedRBACPermissions{
			Roles:       []string{"viewer"},
			Permissions: []string{"user:read"},
		}))
		h := RBACMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &storage.App{}, "machine")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-Organization-Id", oid.String())
		ctx := context.WithValue(context.Background(), types.UserContextKey, u)
		h.ServeHTTP(rec, req.WithContext(ctx))
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("missing org in header and context", func(t *testing.T) {
		h := RBACMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &storage.App{}, "machine")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(context.Background(), types.UserContextKey, u)
		h.ServeHTTP(rec, req.WithContext(ctx))
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func Test_validateCachedPermissions_and_validateAndCachePermissions(t *testing.T) {
	c := testRedisCache(t)
	InitRBACCache(c)
	t.Cleanup(func() { InitRBACCache(nil) })
	uid := uuid.New().String()
	oid := uuid.New().String()

	require.Nil(t, validateCachedPermissions(uid, oid, "x"))

	require.NoError(t, c.SetRBACPermissions(context.Background(), uid, oid, &cache.CachedRBACPermissions{
		Roles:       []string{"unscoped-role"},
		Permissions: []string{"p1"},
	}))
	got := validateCachedPermissions(uid, oid, "p1")
	require.NotNil(t, got)
	require.False(t, *got)

	require.NoError(t, c.SetRBACPermissions(context.Background(), uid, oid, &cache.CachedRBACPermissions{
		Roles:       []string{"owner"},
		Permissions: []string{"machine:read"},
	}))
	got = validateCachedPermissions(uid, oid, "machine:read")
	require.NotNil(t, got)
	require.True(t, *got)

	app := testApp(t)
	ctxUser := &types.User{ID: uuid.MustParse(uid), Email: "e", Name: "n"}
	ctx := context.WithValue(context.Background(), "http_request", httptest.NewRequest(http.MethodGet, "/", nil))

	body := fmt.Sprintf(`[{"userId":"%s","organizationId":"%s","role":["admin"],"user":{"id":"%s"}}]`, uid, oid, uid)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.URL.Path, "list-members")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	ok := validateAndCachePermissions(ctx, ctxUser, oid, "user:read", app)
	require.True(t, ok)

	cacheRBACPermissionsFromMember(uid, oid, &BetterAuthMember{UserID: uid, Role: []interface{}{"viewer"}})
}

func Test_extractOrganizationID(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.Empty(t, extractOrganizationID(rec, req))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Organization-Id", "not-uuid")
	require.Empty(t, extractOrganizationID(rec2, req2))

	oid := uuid.New().String()
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set("X-Organization-Id", oid)
	require.Equal(t, oid, extractOrganizationID(rec3, req3))

	rec4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/", nil)
	cctx := context.WithValue(context.Background(), types.OrganizationIDKey, oid)
	require.Equal(t, oid, extractOrganizationID(rec4, req4.WithContext(cctx)))
}

func Test_FeatureFlagMiddleware(t *testing.T) {
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	oid := uuid.New()
	_, err := db.ExecContext(ctx, `INSERT INTO feature_flags (id, organization_id, feature_name, is_enabled, created_at, updated_at, deleted_at) VALUES (?,?,?,?,?,?,NULL)`,
		uuid.New().String(), oid.String(), "f1", 1, time.Now().UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano))
	require.NoError(t, err)

	c := testRedisCache(t)
	h := FeatureFlagMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, "f1", c)

	t.Run("missing org in context", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid org uuid in context", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		cctx := context.WithValue(context.Background(), types.OrganizationIDKey, "bad")
		h.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("cache hit enabled", func(t *testing.T) {
		require.NoError(t, c.SetFeatureFlag(ctx, oid.String(), "f1", true))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		cctx := context.WithValue(context.Background(), types.OrganizationIDKey, oid.String())
		h.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("cache hit disabled", func(t *testing.T) {
		require.NoError(t, c.SetFeatureFlag(ctx, oid.String(), "f1", false))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		cctx := context.WithValue(context.Background(), types.OrganizationIDKey, oid.String())
		h.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("redis miss uses db and repopulates cache", func(t *testing.T) {
		_ = c.InvalidateFeatureFlag(ctx, oid.String(), "f1")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		cctx := context.WithValue(context.Background(), types.OrganizationIDKey, oid.String())
		h.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("disable cache header", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Disable-Cache", "true")
		cctx := context.WithValue(context.Background(), types.OrganizationIDKey, oid.String())
		h.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("cache get error falls through to db", func(t *testing.T) {
		mr := miniredis.RunT(t)
		cc, err := cache.NewCache("redis://" + mr.Addr())
		require.NoError(t, err)
		mr.Close()
		h2 := FeatureFlagMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, "f1", cc)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		cctx := context.WithValue(context.Background(), types.OrganizationIDKey, oid.String())
		h2.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusOK, rec.Code)
	})
}

func Test_FeatureFlagMiddleware_dbError(t *testing.T) {
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	oid := uuid.New()
	_, err := db.ExecContext(ctx, `DROP TABLE feature_flags`)
	require.NoError(t, err)
	h := FeatureFlagMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, "f1", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	cctx := context.WithValue(context.Background(), types.OrganizationIDKey, oid.String())
	h.ServeHTTP(rec, req.WithContext(cctx))
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func Test_AuditMiddleware_paths(t *testing.T) {
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	uid := uuid.New()
	oid := uuid.New()
	insertUser(t, db, ctx, uid, "audit@x")
	log := logger.NewLogger()

	t.Run("skip no user", func(t *testing.T) {
		h := AuditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), app, log, "user")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("skip no org", func(t *testing.T) {
		h := AuditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), app, log, "user")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		cctx := context.WithValue(context.Background(), types.UserContextKey, &types.User{ID: uid})
		h.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("skip invalid org", func(t *testing.T) {
		h := AuditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), app, log, "user")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		cctx := context.WithValue(context.Background(), types.UserContextKey, &types.User{ID: uid})
		cctx = context.WithValue(cctx, types.OrganizationIDKey, "bad")
		h.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("skip GET", func(t *testing.T) {
		h := AuditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), app, log, "user")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		cctx := context.WithValue(context.Background(), types.UserContextKey, &types.User{ID: uid})
		cctx = context.WithValue(cctx, types.OrganizationIDKey, oid.String())
		h.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("map resource types and endpoint", func(t *testing.T) {
		for _, rt := range []string{
			"user", "organization", "role", "permission", "application", "deploy", "deployment",
			"domain", "github-connector", "smtp", "notification", "feature_flags",
			"container", "audit", "terminal", "integration", "unknown-map-default",
		} {
			_ = mapResourceType(rt)
		}
		require.Equal(t, "machines", getEndpointName("/api/v1/machines/list"))
		require.Equal(t, "unknown", getEndpointName("/api/v1/"))
		id := uuid.New()
		require.Equal(t, id, extractResourceIDFromPath("/api/v1/x/"+id.String()+"/y"))
	})

	t.Run("POST body read error", func(t *testing.T) {
		h := AuditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), app, log, "user")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", errReader{})
		cctx := context.WithValue(context.Background(), types.UserContextKey, &types.User{ID: uid})
		cctx = context.WithValue(cctx, types.OrganizationIDKey, oid.String())
		h.ServeHTTP(rec, req.WithContext(cctx))
		time.Sleep(50 * time.Millisecond)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("POST success writes audit async", func(t *testing.T) {
		h := AuditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}), app, log, "user")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", bytesReader(`{"a":1}`))
		cctx := context.WithValue(context.Background(), types.UserContextKey, &types.User{ID: uid})
		cctx = context.WithValue(cctx, types.OrganizationIDKey, oid.String())
		h.ServeHTTP(rec, req.WithContext(cctx))
		time.Sleep(150 * time.Millisecond)
		require.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("non success skips audit log", func(t *testing.T) {
		h := AuditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}), app, log, "user")
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", bytesReader(`{}`))
		cctx := context.WithValue(context.Background(), types.UserContextKey, &types.User{ID: uid})
		cctx = context.WithValue(cctx, types.OrganizationIDKey, oid.String())
		h.ServeHTTP(rec, req.WithContext(cctx))
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("auditResponseWriter WriteHeader", func(t *testing.T) {
		base := httptest.NewRecorder()
		arw := &auditResponseWriter{ResponseWriter: base, statusCode: 200}
		arw.WriteHeader(http.StatusTeapot)
		require.Equal(t, http.StatusTeapot, arw.statusCode)
	})

	t.Run("getAuditActionFromMethod default", func(t *testing.T) {
		require.Equal(t, types.AuditActionAccess, getAuditActionFromMethod("CUSTOM"))
	})
}

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) { return 0, errors.New("read err") }

func bytesReader(s string) io.ReadCloser { return io.NopCloser(strings.NewReader(s)) }
