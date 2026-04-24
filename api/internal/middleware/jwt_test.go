package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/require"
)

func Test_isJWT(t *testing.T) {
	require.False(t, isJWT("nope"))
	require.False(t, isJWT("a.b"))
	require.True(t, isJWT("a.b.c"))
}

func Test_fetchJWKS_errorsAndCache(t *testing.T) {
	resetJWKSState(t)

	t.Setenv("AUTH_JWKS_URL", "")
	_, err := fetchJWKS(context.Background())
	require.ErrorContains(t, err, "AUTH_JWKS_URL not configured")

	resetJWKSState(t)
	t.Setenv("AUTH_JWKS_URL", "http://127.0.0.1:9/jwks")
	_, err = fetchJWKS(context.Background())
	require.Error(t, err)

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
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		require.NoError(t, enc.Encode(set))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resetJWKSState(t)
	t.Setenv("AUTH_JWKS_URL", srv.URL+"/jwks")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_JWKS_URL") })

	ks1, err := fetchJWKS(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ks1)

	ks2, err := fetchJWKS(context.Background())
	require.NoError(t, err)
	require.Equal(t, ks1, ks2)
}

func Test_jwksIfWarmLocked(t *testing.T) {
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
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(set))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_JWKS_URL", srv.URL+"/jwks")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_JWKS_URL") })

	_, err = fetchJWKS(context.Background())
	require.NoError(t, err)

	jwksMu.Lock()
	got, ok := jwksIfWarmLocked()
	require.True(t, ok)
	require.NotNil(t, got)
	jwksExpiry = time.Now().Add(-time.Hour)
	got2, ok2 := jwksIfWarmLocked()
	require.False(t, ok2)
	require.Nil(t, got2)
	jwksMu.Unlock()
}

func Test_fetchJWKS_postLockWarmReturnDeterministic(t *testing.T) {
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
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(set))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_JWKS_URL", srv.URL+"/jwks")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_JWKS_URL") })

	_, err = fetchJWKS(context.Background())
	require.NoError(t, err)
	jwksMu.Lock()
	jwksExpiry = time.Now().Add(-time.Hour)
	jwksMu.Unlock()

	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	var once sync.Once
	jwksTestingAwaitBeforeWriteLock = func() {
		once.Do(func() {
			close(ch1)
			<-ch2
		})
	}
	t.Cleanup(func() { jwksTestingAwaitBeforeWriteLock = nil })

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := fetchJWKS(context.Background())
		require.NoError(t, err)
	}()

	<-ch1

	go func() {
		defer wg.Done()
		_, err := fetchJWKS(context.Background())
		require.NoError(t, err)
	}()

	time.Sleep(20 * time.Millisecond)
	close(ch2)
	wg.Wait()
}

func Test_validateM2MJWT_whenFetchJWKSError(t *testing.T) {
	resetJWKSState(t)
	t.Setenv("AUTH_JWKS_URL", "")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_JWKS_URL") })
	_, err := validateM2MJWT(context.Background(), "a.b.c", "")
	require.ErrorContains(t, err, "failed to fetch JWKS")
}

func Test_validateM2MJWT_fullClaimPath(t *testing.T) {
	resetJWKSState(t)
	orgID := "33333333-3333-3333-3333-333333333333"

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
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(set)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	t.Setenv("AUTH_JWKS_URL", srv.URL+"/jwks")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_JWKS_URL") })
	t.Setenv("AUTH_ISSUER", "https://issuer.test")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_ISSUER") })
	t.Setenv("AUTH_AUDIENCE", "nixopus-api")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_AUDIENCE") })

	now := time.Now()
	tok, err := jwt.NewBuilder().
		Issuer("https://issuer.test").
		Audience([]string{"nixopus-api"}).
		Subject("m2m").
		IssuedAt(now.Add(-time.Minute)).
		NotBefore(now.Add(-time.Minute)).
		Expiration(now.Add(time.Hour)).
		Claim("https://nixopus.com/org", orgID).
		Build()
	require.NoError(t, err)

	ph := jws.NewHeaders()
	require.NoError(t, ph.Set(jws.KeyIDKey, "kid1"))
	raw, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, key, jws.WithProtectedHeaders(ph)))
	require.NoError(t, err)

	got, err := validateM2MJWT(context.Background(), string(raw), "")
	require.NoError(t, err)
	require.Equal(t, orgID, got)
}

func Test_validateM2MJWT_headerOrgFallback(t *testing.T) {
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
		Build()
	require.NoError(t, err)
	ph := jws.NewHeaders()
	require.NoError(t, ph.Set(jws.KeyIDKey, "kid1"))
	raw, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, key, jws.WithProtectedHeaders(ph)))
	require.NoError(t, err)

	got, err := validateM2MJWT(context.Background(), string(raw), "44444444-4444-4444-4444-444444444444")
	require.NoError(t, err)
	require.Equal(t, "44444444-4444-4444-4444-444444444444", got)
}

func Test_validateM2MJWT_nonStringOrgClaim(t *testing.T) {
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
		Claim("https://nixopus.com/org", 123).
		Build()
	require.NoError(t, err)
	ph := jws.NewHeaders()
	require.NoError(t, ph.Set(jws.KeyIDKey, "kid1"))
	raw, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, key, jws.WithProtectedHeaders(ph)))
	require.NoError(t, err)

	_, err = validateM2MJWT(context.Background(), string(raw), "")
	require.ErrorContains(t, err, "missing organization")
}

func Test_validateM2MJWT_parseFails(t *testing.T) {
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

	_, err = validateM2MJWT(context.Background(), "not.a.jwt", "")
	require.Error(t, err)
}
