package channel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTokenServer returns a test server that serves a valid token response.
// If delay > 0 it sleeps before responding, giving concurrent goroutines time
// to pile up waiting for the write lock inside getToken().
func newTokenServer(t *testing.T, accessToken string, expiresIn int, delay time.Duration) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if delay > 0 {
			time.Sleep(delay)
		}
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
		})
	}))
}

func newWebhookServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}

func baseMsg() Message {
	return Message{
		Body: "hello agent",
		Metadata: map[string]string{
			"event_type":      "test.event",
			"organization_id": "org-1",
			"user_id":         "user-1",
			"extra_key":       "extra_val",
		},
	}
}

// ---------------------------------------------------------------------------
// Constructor + Type
// ---------------------------------------------------------------------------

func TestNewAgentChannel_ReturnsNonNil(t *testing.T) {
	ch := NewAgentChannel("https://wh.example.com", "https://token.example.com", "cid", "csecret")
	require.NotNil(t, ch)
	assert.NotNil(t, ch.httpClient)
}

func TestAgentChannel_Type(t *testing.T) {
	ch := NewAgentChannel("", "", "", "")
	assert.Equal(t, "agent", ch.Type())
}

// ---------------------------------------------------------------------------
// Send — token acquisition failures
// ---------------------------------------------------------------------------

func TestAgentChannel_Send_TokenServer_Non200(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer tokenSrv.Close()

	webhookSrv := newWebhookServer(t, http.StatusOK)
	defer webhookSrv.Close()

	ch := NewAgentChannel(webhookSrv.URL, tokenSrv.URL, "cid", "csecret")

	err := ch.Send(context.Background(), baseMsg())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent channel token acquisition failed")
	assert.Contains(t, err.Error(), "token endpoint returned status 403")
}

func TestAgentChannel_Send_TokenServer_InvalidJSON(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer tokenSrv.Close()

	ch := NewAgentChannel("http://wh.example.com", tokenSrv.URL, "cid", "csecret")

	err := ch.Send(context.Background(), baseMsg())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent channel token acquisition failed")
	assert.Contains(t, err.Error(), "failed to decode token response")
}

func TestAgentChannel_Send_TokenServer_HTTPError(t *testing.T) {
	// Point to a server that is already closed → connection refused.
	closedSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	closedSrv.Close()

	ch := NewAgentChannel("http://wh.example.com", closedSrv.URL, "cid", "csecret")

	err := ch.Send(context.Background(), baseMsg())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent channel token acquisition failed")
}

// ---------------------------------------------------------------------------
// Send — webhook failures (token succeeds)
// ---------------------------------------------------------------------------

func TestAgentChannel_Send_Webhook_HTTPError(t *testing.T) {
	tokenSrv := newTokenServer(t, "tok", 3600, 0)
	defer tokenSrv.Close()

	// Closed webhook server → connection refused.
	webhookSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	webhookSrv.Close()

	ch := NewAgentChannel(webhookSrv.URL, tokenSrv.URL, "cid", "csecret")

	err := ch.Send(context.Background(), baseMsg())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent channel webhook POST failed")
}

func TestAgentChannel_Send_Webhook_Non2xx(t *testing.T) {
	tokenSrv := newTokenServer(t, "tok", 3600, 0)
	defer tokenSrv.Close()

	webhookSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request body", http.StatusBadRequest)
	}))
	defer webhookSrv.Close()

	ch := NewAgentChannel(webhookSrv.URL, tokenSrv.URL, "cid", "csecret")

	err := ch.Send(context.Background(), baseMsg())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent channel webhook returned status 400")
}

// ---------------------------------------------------------------------------
// Send — success
// ---------------------------------------------------------------------------

func TestAgentChannel_Send_Success(t *testing.T) {
	tokenSrv := newTokenServer(t, "mytoken", 3600, 0)
	defer tokenSrv.Close()

	var receivedAuth string
	webhookSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer webhookSrv.Close()

	ch := NewAgentChannel(webhookSrv.URL, tokenSrv.URL, "cid", "csecret")

	err := ch.Send(context.Background(), baseMsg())
	require.NoError(t, err)
	assert.Equal(t, "Bearer mytoken", receivedAuth)
}

// ---------------------------------------------------------------------------
// getToken — cached token path (RLock fast path)
// ---------------------------------------------------------------------------

func TestAgentChannel_getToken_CacheHit(t *testing.T) {
	ch := NewAgentChannel("http://wh.example.com", "http://token.example.com", "cid", "csecret")
	// Seed the cache directly (same package).
	ch.cachedToken = "pre-cached"
	ch.tokenExpiry = time.Now().Add(time.Hour)

	token, err := ch.getToken()
	require.NoError(t, err)
	assert.Equal(t, "pre-cached", token)
}

// ---------------------------------------------------------------------------
// getToken — short-lived token (covers the skew < expiresIn/2 branch)
// ---------------------------------------------------------------------------

func TestAgentChannel_getToken_ShortLivedToken(t *testing.T) {
	// ExpiresIn = 5 seconds → 5s < 2*5min, so skew = ExpiresIn/2 = 2.5s
	tokenSrv := newTokenServer(t, "short-token", 5, 0)
	defer tokenSrv.Close()

	ch := NewAgentChannel("http://wh.example.com", tokenSrv.URL, "cid", "csecret")
	token, err := ch.getToken()
	require.NoError(t, err)
	assert.Equal(t, "short-token", token)
	// Expiry should be within the next 5 seconds (skew = 2.5s → expiry ≤ now+2.5s)
	assert.True(t, ch.tokenExpiry.Before(time.Now().Add(5*time.Second)))
}

// ---------------------------------------------------------------------------
// getToken — http.NewRequest creation error (invalid token URL)
// ---------------------------------------------------------------------------

func TestAgentChannel_getToken_TokenRequestCreationError(t *testing.T) {
	// "://invalid" is not a valid URL — http.NewRequest will fail.
	ch := NewAgentChannel("http://wh.example.com", "://invalid-token-url", "cid", "csecret")

	_, err := ch.getToken()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create token request")
}

// ---------------------------------------------------------------------------
// Send — http.NewRequestWithContext creation error (invalid webhook URL)
// ---------------------------------------------------------------------------

func TestAgentChannel_Send_WebhookRequestCreationError(t *testing.T) {
	// Token server is valid; webhook URL is invalid so request creation fails.
	tokenSrv := newTokenServer(t, "tok", 3600, 0)
	defer tokenSrv.Close()

	ch := NewAgentChannel("://invalid-webhook", tokenSrv.URL, "cid", "csecret")

	err := ch.Send(context.Background(), baseMsg())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent channel request creation failed")
}

// ---------------------------------------------------------------------------
// getToken — double-checked lock (lines 120-122)
//
// Strategy:
//  1. The test holds ch.mu.RLock() — multiple readers can still acquire RLock,
//     but any goroutine trying Lock() will block until ALL readers release.
//  2. N goroutines are released; they each call RLock (succeeds), see cache
//     miss, call RUnlock, then call Lock() — all N block here.
//  3. After a brief sleep (goroutines are now queued at Lock), the test
//     releases its RLock.
//  4. The first goroutine gets the write Lock, fetches the token (slow server),
//     sets cachedToken, then releases Lock.
//  5. Each subsequent goroutine gets Lock one-by-one and hits the double-check
//     at lines 120-122 where cachedToken is already populated.
// ---------------------------------------------------------------------------

func TestAgentChannel_getToken_DoubleCheckLock(t *testing.T) {
	prev := runtime.GOMAXPROCS(4)
	t.Cleanup(func() { runtime.GOMAXPROCS(prev) })

	// Token server with a deliberate delay so the first lock-holder keeps the
	// write lock long enough for other goroutines to queue behind it.
	tokenSrv := newTokenServer(t, "lock-token", 3600, 20*time.Millisecond)
	defer tokenSrv.Close()

	ch := NewAgentChannel("http://wh.example.com", tokenSrv.URL, "cid", "csecret")

	const N = 5
	start := make(chan struct{})
	var wg sync.WaitGroup
	tokens := make([]string, N)
	errs := make([]error, N)

	// Hold a read lock: goroutines can pass RLock (cache miss) but will then
	// block at Lock() until we release this reader lock.
	ch.mu.RLock()

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			tok, err := ch.getToken()
			tokens[idx] = tok
			errs[idx] = err
		}(i)
	}

	close(start)
	// Pure mutex operations take nanoseconds; 5ms is extremely generous.
	time.Sleep(5 * time.Millisecond)

	// Release the read lock — goroutines can now compete for the write lock.
	ch.mu.RUnlock()

	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "goroutine %d failed", i)
		assert.Equal(t, "lock-token", tokens[i])
	}
	assert.Equal(t, "lock-token", ch.cachedToken)
}
