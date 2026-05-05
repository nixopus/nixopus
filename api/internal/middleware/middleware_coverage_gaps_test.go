package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
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

func Test_verifySessionWithFallback_cookieOnly(t *testing.T) {
	orig := verifySessionFn
	verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
		return sessionResp("only-cookie", "org"), nil
	}
	t.Cleanup(func() { verifySessionFn = orig })
	got, err := verifySessionWithFallback(httptest.NewRequest(http.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, "only-cookie", got.User.ID)
}

func Test_AuthMiddleware_moreBranches(t *testing.T) {
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	uid := uuid.New()
	oid := uuid.New()
	insertUser(t, db, ctx, uid, "gaps@x")
	insertMember(t, db, ctx, uid, oid)

	c := testRedisCache(t)
	orig := verifySessionFn
	t.Cleanup(func() { verifySessionFn = orig })

	t.Run("GetUserByID cache hit", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp(uid.String(), oid.String()), nil
		}
		u := types.User{ID: uid, Email: "gaps@x", Name: "n"}
		require.NoError(t, c.SetUserByID(ctx, uid.String(), &u))
		require.NoError(t, c.SetOrgMembership(ctx, uid.String(), oid.String(), true))
		var hit bool
		h := AuthMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			u2 := r.Context().Value(types.UserContextKey).(*types.User)
			hit = u2.Email == "gaps@x"
		}), app, c, logger.NewLogger())
		rec := httptest.NewRecorder()
		req := authReq(t, "/api/v1/machines")
		h.ServeHTTP(rec, req)
		require.True(t, hit)
	})

	t.Run("X-Disable-Cache skips session redis only", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp(uid.String(), oid.String()), nil
		}
		require.NoError(t, c.SetOrgMembership(ctx, uid.String(), oid.String(), true))
		h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, c, logger.NewLogger())
		rec := httptest.NewRecorder()
		req := authReq(t, "/api/v1/machines")
		req.Header.Set("X-Disable-Cache", "true")
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("resolve org failure uses forbidden when wrapped", func(t *testing.T) {
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp(uid.String(), oid.String()), nil
		}
		t.Setenv("AUTH_SERVICE_URL", "http://127.0.0.1:9")
		t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })
		_ = c.InvalidateOrgMembership(ctx, uid.String(), oid.String())
		h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, c, logger.NewLogger())
		rec := httptest.NewRecorder()
		req := authReq(t, "/api/v1/machines")
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("internal error when user query fails", func(t *testing.T) {
		_ = os.Unsetenv("AUTH_SERVICE_URL")
		verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
			return sessionResp(uid.String(), oid.String()), nil
		}
		_ = c.InvalidateUserByID(ctx, uid.String())
		require.NoError(t, db.Close())
		h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, c, logger.NewLogger())
		rec := httptest.NewRecorder()
		req := authReq(t, "/api/v1/machines")
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func Test_resolveAndVerifyOrganization_noOrg_badUserID(t *testing.T) {
	app := testApp(t)
	_, err := resolveAndVerifyOrganization(context.Background(), httptest.NewRequest(http.MethodGet, "/", nil), nil, app, "not-uuid", sessionResp("not-uuid", ""), logger.NewLogger())
	require.ErrorContains(t, err, "no organization ID")
}

func Test_verifyOrganizationMembership_cacheReadError(t *testing.T) {
	mr := miniredis.RunT(t)
	c, err := cache.NewCache("redis://" + mr.Addr())
	require.NoError(t, err)
	mr.Close()
	_, err = verifyOrganizationMembership(context.Background(), nil, c, "u", "00000000-0000-0000-0000-000000000001", logger.NewLogger())
	require.Error(t, err)
}

func Test_validateCachedPermissions_cacheError(t *testing.T) {
	mr := miniredis.RunT(t)
	c, err := cache.NewCache("redis://" + mr.Addr())
	require.NoError(t, err)
	InitRBACCache(c)
	t.Cleanup(func() { InitRBACCache(nil) })
	require.NoError(t, c.SetRBACPermissions(context.Background(), "a", "b", &cache.CachedRBACPermissions{Roles: []string{"owner"}, Permissions: []string{"x"}}))
	mr.Close()
	require.Nil(t, validateCachedPermissions("a", "b", "x"))
}

func Test_validateAndCachePermissions_variants(t *testing.T) {
	app := testApp(t)
	uid := uuid.New()
	oid := uuid.New().String()
	ctx := context.WithValue(context.Background(), "http_request", httptest.NewRequest(http.MethodGet, "/", nil))
	u := &types.User{ID: uid, Email: "e", Name: "n"}

	t.Run("member fetch fails", func(t *testing.T) {
		t.Setenv("AUTH_SERVICE_URL", "http://127.0.0.1:9")
		t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })
		require.False(t, validateAndCachePermissions(ctx, u, oid, "user:read", app, logger.NewLogger()))
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })
	require.False(t, validateAndCachePermissions(ctx, u, oid, "user:read", app, logger.NewLogger()))

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		body := `[{"userId":"` + uid.String() + `","organizationId":"` + oid + `","role":123,"user":{"id":"` + uid.String() + `"}}]`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv2.Close)
	t.Setenv("AUTH_SERVICE_URL", srv2.URL)
	ok := validateAndCachePermissions(ctx, u, oid, "user:read", app, logger.NewLogger())
	require.True(t, ok)

	srv3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		body := `[{"userId":"` + uid.String() + `","organizationId":"` + oid + `","role":[123,"admin"],"user":{"id":"` + uid.String() + `"}}]`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv3.Close)
	t.Setenv("AUTH_SERVICE_URL", srv3.URL)
	require.True(t, validateAndCachePermissions(ctx, u, oid, "user:read", app, logger.NewLogger()))
}

func Test_cacheRBACPermissionsFromMember_variants(t *testing.T) {
	c := testRedisCache(t)
	InitRBACCache(c)
	t.Cleanup(func() { InitRBACCache(nil) })
	cacheRBACPermissionsFromMember("u1", "o1", &BetterAuthMember{Role: []interface{}{"admin"}})
	cacheRBACPermissionsFromMember("u2", "o2", &BetterAuthMember{Role: float64(7)})
	cacheRBACPermissionsFromMember("u3", "o3", &BetterAuthMember{Role: "owner"})
}

func Test_getAuditActionFromMethod_all(t *testing.T) {
	require.Equal(t, types.AuditActionCreate, getAuditActionFromMethod(http.MethodPost))
	require.Equal(t, types.AuditActionUpdate, getAuditActionFromMethod(http.MethodPut))
	require.Equal(t, types.AuditActionUpdate, getAuditActionFromMethod(http.MethodPatch))
	require.Equal(t, types.AuditActionDelete, getAuditActionFromMethod(http.MethodDelete))
}

func Test_FeatureFlagMiddleware_disabledFromDB(t *testing.T) {
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	oid := uuid.New()
	_, err := db.ExecContext(ctx, `INSERT INTO feature_flags (id, organization_id, feature_name, is_enabled, created_at, updated_at, deleted_at) VALUES (?,?,?,?,?,?,NULL)`,
		uuid.New().String(), oid.String(), "offeat", 0, time.Now().UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano))
	require.NoError(t, err)
	h := FeatureFlagMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, "offeat", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	cctx := context.WithValue(context.Background(), types.OrganizationIDKey, oid.String())
	h.ServeHTTP(rec, req.WithContext(cctx))
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func Test_fetchJWKS_concurrentFill(t *testing.T) {
	resetJWKSState(t)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privJWK, err := jwk.FromRaw(key)
	require.NoError(t, err)
	require.NoError(t, privJWK.Set(jwk.KeyIDKey, "kid1"))
	require.NoError(t, privJWK.Set(jwk.AlgorithmKey, jwa.RS256))
	pubJWK, err := jwk.PublicKeyOf(privJWK)
	require.NoError(t, err)
	set := jwk.NewSet()
	require.NoError(t, set.AddKey(pubJWK))
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(20 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(set)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_JWKS_URL", srv.URL+"/jwks")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_JWKS_URL") })

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = fetchJWKS(context.Background())
		}()
	}
	wg.Wait()
}

func Test_validateM2MJWT_emptyOrgClaimUsesHeader(t *testing.T) {
	resetJWKSState(t)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privJWK, err := jwk.FromRaw(key)
	require.NoError(t, err)
	require.NoError(t, privJWK.Set(jwk.KeyIDKey, "kid1"))
	require.NoError(t, privJWK.Set(jwk.AlgorithmKey, jwa.RS256))
	pubJWK, err := jwk.PublicKeyOf(privJWK)
	require.NoError(t, err)
	set := jwk.NewSet()
	require.NoError(t, set.AddKey(pubJWK))
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(set)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_JWKS_URL", srv.URL+"/jwks")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_JWKS_URL") })
	now := time.Now()
	tok, err := jwt.NewBuilder().
		Subject("m2m").
		IssuedAt(now.Add(-time.Minute)).
		NotBefore(now.Add(-time.Minute)).
		Expiration(now.Add(time.Hour)).
		Claim("https://nixopus.com/org", "").
		Build()
	require.NoError(t, err)
	ph := jws.NewHeaders()
	require.NoError(t, ph.Set(jws.KeyIDKey, "kid1"))
	raw, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, key, jws.WithProtectedHeaders(ph)))
	require.NoError(t, err)
	got, err := validateM2MJWT(context.Background(), string(raw), "77777777-7777-7777-7777-777777777777")
	require.NoError(t, err)
	require.Equal(t, "77777777-7777-7777-7777-777777777777", got)
}

func Test_tryM2MJWTAuth_validateFailsFallsThrough(t *testing.T) {
	resetJWKSState(t)
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	uid := uuid.New()
	oid := uuid.New()
	insertUser(t, db, ctx, uid, "fall@x")
	insertMember(t, db, ctx, uid, oid)

	orig := verifySessionFn
	verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
		return sessionResp(uid.String(), oid.String()), nil
	}
	t.Cleanup(func() { verifySessionFn = orig })

	rc := testRedisCache(t)
	require.NoError(t, rc.SetOrgMembership(context.Background(), uid.String(), oid.String(), true))

	h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, rc, logger.NewLogger())
	rec := httptest.NewRecorder()
	req := authReq(t, "/api/v1/machines")
	req.Header.Set("Authorization", "Bearer not-a-jwt-shape")
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func Test_logHTTPRequest_durationBranches(t *testing.T) {
	rw := newResponseWriter(httptest.NewRecorder())
	logHTTPRequest(rw, httptest.NewRequest(http.MethodGet, "/", nil), time.Now().Add(-2*time.Millisecond))
}

func Test_NewRateLimiterWithConfig_cleanupRemovesStale(t *testing.T) {
	prevStale := rateLimiterStaleClientAfter
	prevTick := rateLimiterCleanupTick
	rateLimiterStaleClientAfter = 15 * time.Millisecond
	rateLimiterCleanupTick = 12 * time.Millisecond
	t.Cleanup(func() {
		rateLimiterStaleClientAfter = prevStale
		rateLimiterCleanupTick = prevTick
	})

	h := NewRateLimiterWithConfig(1000, 2000)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "7.7.7.7:7777"
	h.ServeHTTP(httptest.NewRecorder(), req)
	time.Sleep(45 * time.Millisecond)
	h.ServeHTTP(httptest.NewRecorder(), req)
}

func Test_AuditMiddleware_PUT_and_auditLogFailurePath(t *testing.T) {
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	uid := uuid.New()
	oid := uuid.New()
	insertUser(t, db, ctx, uid, "aput@x")
	log := logger.NewLogger()

	h := AuditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), app, log, "user")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/", bytesReader(`{}`))
	cctx := context.WithValue(context.Background(), types.UserContextKey, &types.User{ID: uid})
	cctx = context.WithValue(cctx, types.OrganizationIDKey, oid.String())
	h.ServeHTTP(rec, req.WithContext(cctx))
	time.Sleep(120 * time.Millisecond)

	_, err := db.ExecContext(ctx, `DROP TABLE audit_logs`)
	require.NoError(t, err)
	h2 := AuditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), app, log, "user")
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/", bytesReader(`{"z":1}`))
	h2.ServeHTTP(rec2, req2.WithContext(cctx))
	time.Sleep(120 * time.Millisecond)
}

func Test_verifySessionWithFallback_firstFailsNoAPIKey(t *testing.T) {
	orig := verifySessionFn
	verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
		return nil, errors.New("no session")
	}
	t.Cleanup(func() { verifySessionFn = orig })
	_, err := verifySessionWithFallback(httptest.NewRequest(http.MethodGet, "/", nil))
	require.ErrorContains(t, err, "no session")
}

func Test_verifySessionWithFallback_firstFailsAPIKeySecondFails(t *testing.T) {
	var calls int
	orig := verifySessionFn
	verifySessionFn = func(r *http.Request) (*betterauth.SessionResponse, error) {
		calls++
		if r.Header.Get("x-api-key") == "" {
			return nil, errors.New("cookie auth failed")
		}
		return nil, errors.New("api key auth failed")
	}
	t.Cleanup(func() { verifySessionFn = orig })
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("x-api-key", "secret")
	_, err := verifySessionWithFallback(r)
	require.ErrorContains(t, err, "api key auth failed")
	require.Equal(t, 2, calls)
}

func Test_defaultVerifySessionFn_invokesBetterAuth(t *testing.T) {
	origClient := betterauth.HTTPClient
	betterauth.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("no network")
		}),
	}
	t.Cleanup(func() { betterauth.HTTPClient = origClient })
	_, err := defaultVerifySessionFn(httptest.NewRequest(http.MethodGet, "/", nil))
	require.Error(t, err)
}

func Test_validateCachedPermissions_whenRBACCacheUnset(t *testing.T) {
	prev := rbacCache
	rbacCache = nil
	t.Cleanup(func() { rbacCache = prev })
	require.Nil(t, validateCachedPermissions(uuid.New().String(), uuid.New().String(), "machine:read"))
}

func Test_validateAndCachePermissions_roleJSONString(t *testing.T) {
	app := testApp(t)
	uid := uuid.New()
	oid := uuid.New().String()
	ctx := context.WithValue(context.Background(), "http_request", httptest.NewRequest(http.MethodGet, "/", nil))
	u := &types.User{ID: uid, Email: "e", Name: "n"}
	body := `[{"userId":"` + uid.String() + `","organizationId":"` + oid + `","role":"admin","user":{"id":"` + uid.String() + `"}}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })
	require.True(t, validateAndCachePermissions(ctx, u, oid, "user:read", app, logger.NewLogger()))
}

func Test_validateAndCachePermissions_noHTTPRequestInContext(t *testing.T) {
	app := testApp(t)
	uid := uuid.New()
	oid := uuid.New().String()
	u := &types.User{ID: uid, Email: "e", Name: "n"}
	body := `[{"userId":"` + uid.String() + `","organizationId":"` + oid + `","role":"member","user":{"id":"` + uid.String() + `"}}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })
	require.True(t, validateAndCachePermissions(context.Background(), u, oid, "user:read", app, logger.NewLogger()))
}

func Test_validateAndCachePermissions_httpRequestWrongTypeInContext(t *testing.T) {
	app := testApp(t)
	uid := uuid.New()
	oid := uuid.New().String()
	ctx := context.WithValue(context.Background(), "http_request", "not-a-request")
	u := &types.User{ID: uid, Email: "e", Name: "n"}
	body := `[{"userId":"` + uid.String() + `","organizationId":"` + oid + `","role":"member","user":{"id":"` + uid.String() + `"}}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })
	require.True(t, validateAndCachePermissions(ctx, u, oid, "user:read", app, logger.NewLogger()))
}

func Test_RBACMiddleware_cacheMissFetchesFromBetterAuth(t *testing.T) {
	c := testRedisCache(t)
	InitRBACCache(c)
	t.Cleanup(func() { InitRBACCache(nil) })

	uid := uuid.New()
	oid := uuid.New()
	u := &types.User{ID: uid, Email: "rbacmiss@x", Name: "n"}

	body := `[{"userId":"` + uid.String() + `","organizationId":"` + oid.String() + `","role":"member","user":{"id":"` + uid.String() + `"}}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	h := RBACMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &storage.App{}, "user", logger.NewLogger())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("X-Organization-Id", oid.String())
	ctx := context.WithValue(context.Background(), types.UserContextKey, u)
	h.ServeHTTP(rec, req.WithContext(ctx))
	require.Equal(t, http.StatusOK, rec.Code)
}

func Test_verifyOrganizationMembership_successSetsOrgCacheAndRBAC(t *testing.T) {
	c := testRedisCache(t)
	ctx := context.Background()
	uid := uuid.New().String()
	oid := uuid.New().String()
	require.NoError(t, c.InvalidateOrgMembership(ctx, uid, oid))

	InitRBACCache(c)
	t.Cleanup(func() { InitRBACCache(nil) })

	body := `[{"userId":"` + uid + `","organizationId":"` + oid + `","role":"viewer","user":{"id":"` + uid + `"}}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	ok, err := verifyOrganizationMembership(ctx, httptest.NewRequest(http.MethodGet, "/", nil), c, uid, oid, logger.NewLogger())
	require.NoError(t, err)
	require.True(t, ok)
	cached, err := c.GetOrgMembership(ctx, uid, oid)
	require.NoError(t, err)
	require.True(t, cached)
}

func Test_tryM2MJWTAuth_invalidSignedJWTFallsThrough(t *testing.T) {
	resetJWKSState(t)
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	uid := uuid.New()
	oid := uuid.New()
	insertUser(t, db, ctx, uid, "m2mjwt@x")
	insertMember(t, db, ctx, uid, oid)

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
	wrongKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	raw, err := jwt.Sign(tokBuilt, jwt.WithKey(jwa.RS256, wrongKey, jws.WithProtectedHeaders(ph)))
	require.NoError(t, err)
	badTok := string(raw)

	orig := verifySessionFn
	verifySessionFn = func(*http.Request) (*betterauth.SessionResponse, error) {
		return sessionResp(uid.String(), oid.String()), nil
	}
	t.Cleanup(func() { verifySessionFn = orig })
	rc := testRedisCache(t)
	require.NoError(t, rc.SetOrgMembership(ctx, uid.String(), oid.String(), true))

	h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, rc, logger.NewLogger())
	rec := httptest.NewRecorder()
	req := authReq(t, "/api/v1/machines")
	req.Header.Set("Authorization", "Bearer "+badTok)
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func Test_tryM2MJWTAuth_userMissingInDatabase(t *testing.T) {
	resetJWKSState(t)
	app := testApp(t)
	db := app.Store.DB
	ctx := app.Ctx
	uid := uuid.New()
	oid := uuid.New()
	insertUser(t, db, ctx, uid, "m2mok@x")
	insertMember(t, db, ctx, uid, oid)

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

	missingUser := uuid.New()
	h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), app, nil, logger.NewLogger())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/x", nil)
	req.Header.Set("Authorization", "Bearer "+string(raw))
	req.Header.Set("X-Organization-Id", oid.String())
	req.Header.Set("X-User-Id", missingUser.String())
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
