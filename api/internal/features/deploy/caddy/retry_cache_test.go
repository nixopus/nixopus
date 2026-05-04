package caddy

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/require"
)

func TestWithRetry_successFirstAttempt(t *testing.T) {
	var calls int
	l := logger.NewLogger()
	err := WithRetry(func() error {
		calls++
		return nil
	}, 3, l, func() {})
	require.NoError(t, err)
	require.Equal(t, 1, calls)
}

func TestWithRetry_invokesOnFailureThenSucceeds(t *testing.T) {
	var failures, onFails int
	l := logger.NewLogger()
	err := WithRetry(func() error {
		failures++
		if failures == 1 {
			return errors.New("transient")
		}
		return nil
	}, 3, l, func() { onFails++ })
	require.NoError(t, err)
	require.Equal(t, 2, failures)
	require.Equal(t, 1, onFails)
}

func TestWithRetry_exhaustsAttempts(t *testing.T) {
	var onFails int
	l := logger.NewLogger()
	permanentErr := errors.New("perm")
	err := WithRetry(func() error { return permanentErr }, 2, l, func() { onFails++ })
	require.ErrorIs(t, err, permanentErr)
	require.Equal(t, 1, onFails)
}

func TestInvalidateTunnelByKey_removesCacheEntryAndClosesListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	tunnel := &CaddyTunnel{
		listener: ln,
		cleanup: func() error {
			return ln.Close()
		},
	}
	key := "unit-test-host.invalid:2019"
	caddyTunnelCacheMu.Lock()
	caddyTunnelCache[key] = &caddyTunnelEntry{
		tunnel:   tunnel,
		client:   nil,
		lastUsed: time.Now(),
	}
	caddyTunnelCacheMu.Unlock()

	t.Cleanup(func() {
		caddyTunnelCacheMu.Lock()
		delete(caddyTunnelCache, key)
		caddyTunnelCacheMu.Unlock()
	})

	InvalidateTunnelByKey(key)

	caddyTunnelCacheMu.RLock()
	_, exists := caddyTunnelCache[key]
	caddyTunnelCacheMu.RUnlock()
	require.False(t, exists)

	conn, err := net.DialTimeout("tcp", ln.Addr().String(), 50*time.Millisecond)
	if err == nil {
		conn.Close()
		t.Fatal("listener should have been closed")
	}
}

func Test_evictIdleTunnels_removesStaleEntry(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	key := "idle-evict.invalid:2019"
	caddyTunnelCacheMu.Lock()
	caddyTunnelCache[key] = &caddyTunnelEntry{
		tunnel: &CaddyTunnel{
			listener: ln,
			cleanup: func() error {
				return ln.Close()
			},
		},
		lastUsed: time.Now().Add(-tunnelIdleTTL - time.Minute),
	}
	caddyTunnelCacheMu.Unlock()

	t.Cleanup(func() {
		caddyTunnelCacheMu.Lock()
		if e, ok := caddyTunnelCache[key]; ok {
			if e.tunnel != nil {
				_ = e.tunnel.Close()
			}
			delete(caddyTunnelCache, key)
		}
		caddyTunnelCacheMu.Unlock()
	})

	evictIdleTunnels()

	caddyTunnelCacheMu.RLock()
	_, exists := caddyTunnelCache[key]
	caddyTunnelCacheMu.RUnlock()
	require.False(t, exists)
}
