package service

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nixopus/nixopus/api/internal/redisclient"
)

const (
	adminRegisteredKey = "auth:admin_registered"

	// Once admin is registered the value is permanent, cache aggressively.
	adminRegisteredTrueTTL = 24 * time.Hour
	// Before first signup, keep a short TTL so the transition is detected quickly.
	adminRegisteredFalseTTL = 30 * time.Second
)

type AuthCache struct {
	client *redis.Client
}

func NewAuthCache(redisURL string) (*AuthCache, error) {
	client, err := redisclient.New(redisURL)
	if err != nil {
		return nil, err
	}
	return &AuthCache{client: client}, nil
}

// GetAdminRegistered returns the cached value and whether a cache hit occurred.
func (c *AuthCache) GetAdminRegistered(ctx context.Context) (registered bool, hit bool, err error) {
	val, err := c.client.Get(ctx, adminRegisteredKey).Result()
	if err == redis.Nil {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return val == "true", true, nil
}

// SetAdminRegistered caches the result with a TTL that depends on the value.
// true  -> long TTL  (state is permanent once an admin exists)
// false -> short TTL (we want to re-check soon so signup is detected quickly)
func (c *AuthCache) SetAdminRegistered(ctx context.Context, registered bool) error {
	val := "false"
	ttl := adminRegisteredFalseTTL
	if registered {
		val = "true"
		ttl = adminRegisteredTrueTTL
	}
	return c.client.Set(ctx, adminRegisteredKey, val, ttl).Err()
}

// InvalidateAdminRegistered removes the cached admin registration status.
func (c *AuthCache) InvalidateAdminRegistered(ctx context.Context) error {
	return c.client.Del(ctx, adminRegisteredKey).Err()
}
