package caddy

import (
	"fmt"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

const (
	DefaultMaxRetries = 3
	DefaultBaseDelay  = 1 * time.Second
	DefaultMaxDelay   = 10 * time.Second
)

// WithRetry executes an operation with exponential backoff. On failure it
// optionally calls onFailure (e.g. to invalidate a stale tunnel) before
// the next attempt.
func WithRetry(op func() error, maxRetries int, lgr logger.Logger, onFailure func()) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := op(); err != nil {
			lastErr = err

			if i == maxRetries-1 {
				break
			}

			backoff := DefaultBaseDelay * time.Duration(1<<uint(i))
			if backoff > DefaultMaxDelay {
				backoff = DefaultMaxDelay
			}

			lgr.Log(logger.Warning, fmt.Sprintf("caddy operation failed (attempt %d/%d), retrying in %s", i+1, maxRetries, backoff), err.Error())

			if onFailure != nil {
				onFailure()
			}

			time.Sleep(backoff)
			continue
		}
		return nil
	}
	return fmt.Errorf("caddy operation failed after %d attempts: %w", maxRetries, lastErr)
}

// InvalidateTunnel removes a cached Caddy tunnel entry for the given host
// so the next GetCaddyClient call creates a fresh connection.
func InvalidateTunnel(host string) {
	port, err := parseCaddyEndpointPort()
	if err != nil {
		return
	}
	key := host + ":" + port

	caddyTunnelCacheMu.Lock()
	defer caddyTunnelCacheMu.Unlock()

	if entry, exists := caddyTunnelCache[key]; exists {
		if entry.tunnel != nil {
			entry.tunnel.Close()
		}
		delete(caddyTunnelCache, key)
	}
}

// InvalidateTunnelByKey removes a cached tunnel entry by its full key (host:port).
func InvalidateTunnelByKey(key string) {
	caddyTunnelCacheMu.Lock()
	defer caddyTunnelCacheMu.Unlock()

	if entry, exists := caddyTunnelCache[key]; exists {
		if entry.tunnel != nil {
			entry.tunnel.Close()
		}
		delete(caddyTunnelCache, key)
	}
}
