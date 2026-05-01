package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func testAuthCache(t *testing.T) (*AuthCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	c, err := NewAuthCache("redis://" + mr.Addr())
	require.NoError(t, err)
	return c, mr
}

func TestNewAuthCache_invalidURL(t *testing.T) {
	_, err := NewAuthCache("not-a-valid-redis-url")
	require.Error(t, err)
}

func TestAuthCache_adminRegistered_missHitSetInvalidate(t *testing.T) {
	c, mr := testAuthCache(t)
	ctx := context.Background()

	v, hit, err := c.GetAdminRegistered(ctx)
	require.NoError(t, err)
	require.False(t, hit)
	require.False(t, v)

	require.NoError(t, c.SetAdminRegistered(ctx, false))
	got, hit, err := c.GetAdminRegistered(ctx)
	require.NoError(t, err)
	require.True(t, hit)
	require.False(t, got)

	ttlFalse := mr.TTL("auth:admin_registered")
	require.Positive(t, ttlFalse)

	require.NoError(t, c.InvalidateAdminRegistered(ctx))
	require.NoError(t, c.SetAdminRegistered(ctx, true))
	got, hit, err = c.GetAdminRegistered(ctx)
	require.NoError(t, err)
	require.True(t, hit)
	require.True(t, got)

	ttlTrue := mr.TTL("auth:admin_registered")
	require.Greater(t, ttlTrue, time.Minute)
}

func TestAuthCache_GetAdminRegistered_redisError(t *testing.T) {
	c, mr := testAuthCache(t)
	ctx := context.Background()
	require.NoError(t, c.SetAdminRegistered(ctx, true))
	mr.Close()

	_, _, err := c.GetAdminRegistered(ctx)
	require.Error(t, err)
}

func TestAuthCache_SetAdminRegistered_redisError(t *testing.T) {
	c, mr := testAuthCache(t)
	ctx := context.Background()
	mr.Close()

	err := c.SetAdminRegistered(ctx, true)
	require.Error(t, err)
}

func TestAuthCache_InvalidateAdminRegistered_redisError(t *testing.T) {
	c, mr := testAuthCache(t)
	ctx := context.Background()
	mr.Close()

	err := c.InvalidateAdminRegistered(ctx)
	require.Error(t, err)
}
