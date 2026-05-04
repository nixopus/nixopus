package caddy

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) {
	t.Helper()
	mr := miniredis.RunT(t)
	opt, err := redis.ParseURL("redis://" + mr.Addr())
	require.NoError(t, err)
	rdb := redis.NewClient(opt)
	t.Cleanup(func() {
		_ = rdb.Close()
		mr.Close()
		queue.Init(nil)
	})

	queue.Init(rdb)
}

func TestTrackExtensionDomain_Untrack_GetExtensionDomains(t *testing.T) {
	setupTestRedis(t)
	org := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	require.NoError(t, TrackExtensionDomain(org, "ext.example.com", "127.0.0.1:3000"))

	routes, err := GetExtensionDomains(context.Background(), org)
	require.NoError(t, err)
	require.Len(t, routes, 1)
	assert.Equal(t, "ext.example.com", routes[0].Domain)
	assert.Equal(t, "127.0.0.1:3000", routes[0].UpstreamDial)

	require.NoError(t, UntrackExtensionDomain(org, "ext.example.com"))
	routes, err = GetExtensionDomains(context.Background(), org)
	require.NoError(t, err)
	assert.Empty(t, routes)
}

func TestExtensionDomain_redis_uninitialized(t *testing.T) {
	queue.Init(nil)
	org := uuid.New()

	assert.Error(t, TrackExtensionDomain(org, "d", "h:1"))
	assert.Error(t, UntrackExtensionDomain(org, "d"))

	_, err := GetExtensionDomains(context.Background(), org)
	assert.Error(t, err)
}

func TestPendingRemoval_roundTrip(t *testing.T) {
	setupTestRedis(t)
	org := uuid.MustParse("22222222-3333-4444-5555-666666666666")

	assert.NoError(t, EnqueuePendingRemoval(org))
	assert.NoError(t, EnqueuePendingRemoval(org, "a.example", "b.example"))

	names, err := GetPendingRemovals(context.Background(), org)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a.example", "b.example"}, names)

	require.NoError(t, ClearPendingRemoval(org, "a.example"))
	names, err = GetPendingRemovals(context.Background(), org)
	require.NoError(t, err)
	assert.Equal(t, []string{"b.example"}, names)
}

func TestPendingRemoval_redis_uninitialized_and_emptyAdds(t *testing.T) {
	queue.Init(nil)
	org := uuid.New()

	assert.Error(t, EnqueuePendingRemoval(org, "x"))
	assert.Error(t, ClearPendingRemoval(org, "x"))
	_, err := GetPendingRemovals(context.Background(), org)
	assert.Error(t, err)

	setupTestRedis(t)
	assert.NoError(t, EnqueuePendingRemoval(org))
	assert.NoError(t, ClearPendingRemoval(org))
}
