package cache

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
)

func swapJSONMarshal(t *testing.T, fn func(any) ([]byte, error)) {
	t.Helper()
	orig := jsonMarshal
	jsonMarshal = fn
	t.Cleanup(func() { jsonMarshal = orig })
}

func testCache(t *testing.T) (*Cache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	c, err := NewCache("redis://" + mr.Addr())
	require.NoError(t, err)
	return c, mr
}

func TestNewCache_invalidURL(t *testing.T) {
	_, err := NewCache("not-a-valid-redis-url")
	require.Error(t, err)
}

func TestNewCache_ok(t *testing.T) {
	_, _ = testCache(t)
}

func sampleUser() types.User {
	return types.User{
		ID:            uuid.New(),
		Name:          "n",
		Email:         "e@x",
		EmailVerified: true,
		IsOnboarded:   false,
		CreatedAt:     time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt:     time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
	}
}

func TestGetUser_miss(t *testing.T) {
	c, _ := testCache(t)
	ctx := context.Background()
	u, err := c.GetUser(ctx, "missing@x")
	require.NoError(t, err)
	require.Nil(t, u)
}

func TestGetUser_hit(t *testing.T) {
	c, _ := testCache(t)
	ctx := context.Background()
	want := sampleUser()
	require.NoError(t, c.SetUser(ctx, want.Email, &want))
	got, err := c.GetUser(ctx, want.Email)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, want.ID, got.ID)
	require.Equal(t, want.Email, got.Email)
}

func TestGetUser_invalidJSON_deletesKey(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	email := "bad@json"
	mr.Set(UserCacheKeyPrefix+email, "not-json")
	_, err := c.GetUser(ctx, email)
	require.Error(t, err)
	_, err = mr.Get(UserCacheKeyPrefix + email)
	require.Error(t, err)
}

func TestGetUser_nilUUIDAfterUnmarshal(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	email := "nilid@x"
	payload := map[string]any{
		"id":             uuid.Nil.String(),
		"name":           "n",
		"email":          email,
		"email_verified": false,
		"is_onboarded":   false,
		"created_at":     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
		"updated_at":     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
	}
	b, err := json.Marshal(payload)
	require.NoError(t, err)
	mr.Set(UserCacheKeyPrefix+email, string(b))
	u, err := c.GetUser(ctx, email)
	require.NoError(t, err)
	require.Nil(t, u)
	_, err = mr.Get(UserCacheKeyPrefix + email)
	require.Error(t, err)
}

func TestGetUser_redisError(t *testing.T) {
	c, mr := testCache(t)
	mr.Close()
	_, err := c.GetUser(context.Background(), "a@b")
	require.Error(t, err)
}

func TestSetUser_nilID_noop(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	u := sampleUser()
	u.ID = uuid.Nil
	require.NoError(t, c.SetUser(ctx, u.Email, &u))
	_, err := mr.Get(UserCacheKeyPrefix + u.Email)
	require.Error(t, err)
}

func TestSetUser_marshalError(t *testing.T) {
	c, _ := testCache(t)
	swapJSONMarshal(t, func(any) ([]byte, error) { return nil, errors.New("marshal") })
	u := sampleUser()
	require.Error(t, c.SetUser(context.Background(), u.Email, &u))
}

func TestSetUser_roundTrip(t *testing.T) {
	c, _ := testCache(t)
	ctx := context.Background()
	u := sampleUser()
	require.NoError(t, c.SetUser(ctx, u.Email, &u))
}

func TestSetUser_redisError(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	u := sampleUser()
	mr.Close()
	require.Error(t, c.SetUser(ctx, u.Email, &u))
}

func TestGetUserByID_miss(t *testing.T) {
	c, _ := testCache(t)
	u, err := c.GetUserByID(context.Background(), uuid.New().String())
	require.NoError(t, err)
	require.Nil(t, u)
}

func TestGetUserByID_hit(t *testing.T) {
	c, _ := testCache(t)
	ctx := context.Background()
	u := sampleUser()
	require.NoError(t, c.SetUserByID(ctx, u.ID.String(), &u))
	got, err := c.GetUserByID(ctx, u.ID.String())
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, u.ID, got.ID)
}

func TestGetUserByID_invalidJSON_deletesKey(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	id := uuid.New().String()
	key := UserByIDCacheKeyPrefix + id
	mr.Set(key, "not-json")
	_, err := c.GetUserByID(ctx, id)
	require.Error(t, err)
	_, err = mr.Get(key)
	require.Error(t, err)
}

func TestGetUserByID_nilUUID_deletesKey(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	id := uuid.New().String()
	key := UserByIDCacheKeyPrefix + id
	payload := map[string]any{
		"id":             uuid.Nil.String(),
		"name":           "n",
		"email":          "e@x",
		"email_verified": false,
		"is_onboarded":   false,
		"created_at":     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
		"updated_at":     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
	}
	b, err := json.Marshal(payload)
	require.NoError(t, err)
	mr.Set(key, string(b))
	u, err := c.GetUserByID(ctx, id)
	require.NoError(t, err)
	require.Nil(t, u)
	_, err = mr.Get(key)
	require.Error(t, err)
}

func TestGetUserByID_redisError(t *testing.T) {
	c, mr := testCache(t)
	mr.Close()
	_, err := c.GetUserByID(context.Background(), uuid.New().String())
	require.Error(t, err)
}

func TestSetUserByID_nilUser(t *testing.T) {
	c, _ := testCache(t)
	require.NoError(t, c.SetUserByID(context.Background(), uuid.New().String(), nil))
}

func TestSetUserByID_nilUUID(t *testing.T) {
	c, _ := testCache(t)
	u := sampleUser()
	u.ID = uuid.Nil
	require.NoError(t, c.SetUserByID(context.Background(), uuid.New().String(), &u))
}

func TestSetUserByID_marshalError(t *testing.T) {
	c, _ := testCache(t)
	swapJSONMarshal(t, func(any) ([]byte, error) { return nil, errors.New("marshal") })
	u := sampleUser()
	require.Error(t, c.SetUserByID(context.Background(), u.ID.String(), &u))
}

func TestSetUserByID_roundTrip(t *testing.T) {
	c, _ := testCache(t)
	ctx := context.Background()
	u := sampleUser()
	require.NoError(t, c.SetUserByID(ctx, u.ID.String(), &u))
}

func TestSetUserByID_redisError(t *testing.T) {
	c, mr := testCache(t)
	mr.Close()
	u := sampleUser()
	require.Error(t, c.SetUserByID(context.Background(), u.ID.String(), &u))
}

func TestOrgMembership_miss(t *testing.T) {
	c, _ := testCache(t)
	ok, err := c.GetOrgMembership(context.Background(), "u1", "o1")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestOrgMembership_true_false(t *testing.T) {
	c, _ := testCache(t)
	ctx := context.Background()
	require.NoError(t, c.SetOrgMembership(ctx, "u1", "o1", true))
	ok, err := c.GetOrgMembership(ctx, "u1", "o1")
	require.NoError(t, err)
	require.True(t, ok)

	require.NoError(t, c.SetOrgMembership(ctx, "u1", "o1", false))
	ok, err = c.GetOrgMembership(ctx, "u1", "o1")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestGetOrgMembership_redisError(t *testing.T) {
	c, mr := testCache(t)
	mr.Close()
	_, err := c.GetOrgMembership(context.Background(), "u", "o")
	require.Error(t, err)
}

func TestInvalidateUser_andInvalidateUserByID_andInvalidateOrgMembership(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	u := sampleUser()
	require.NoError(t, c.SetUser(ctx, u.Email, &u))
	require.NoError(t, c.SetUserByID(ctx, u.ID.String(), &u))
	require.NoError(t, c.SetOrgMembership(ctx, u.ID.String(), "org1", true))

	require.NoError(t, c.InvalidateUser(ctx, u.Email))
	_, err := mr.Get(UserCacheKeyPrefix + u.Email)
	require.Error(t, err)

	require.NoError(t, c.InvalidateUserByID(ctx, u.ID.String()))
	_, err = mr.Get(UserByIDCacheKeyPrefix + u.ID.String())
	require.Error(t, err)

	require.NoError(t, c.InvalidateOrgMembership(ctx, u.ID.String(), "org1"))
	_, err = mr.Get(OrgMembershipCacheKeyPrefix + u.ID.String() + ":org1")
	require.Error(t, err)
}

func TestFeatureFlag_miss(t *testing.T) {
	c, _ := testCache(t)
	enabled, err := c.GetFeatureFlag(context.Background(), "org", "feat")
	require.ErrorIs(t, err, redis.Nil)
	require.False(t, enabled)
}

func TestFeatureFlag_hit_true_false(t *testing.T) {
	c, _ := testCache(t)
	ctx := context.Background()
	require.NoError(t, c.SetFeatureFlag(ctx, "org", "f", true))
	ok, err := c.GetFeatureFlag(ctx, "org", "f")
	require.NoError(t, err)
	require.True(t, ok)

	require.NoError(t, c.SetFeatureFlag(ctx, "org", "f", false))
	ok, err = c.GetFeatureFlag(ctx, "org", "f")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestGetFeatureFlag_redisError(t *testing.T) {
	c, mr := testCache(t)
	mr.Close()
	_, err := c.GetFeatureFlag(context.Background(), "o", "f")
	require.Error(t, err)
	require.False(t, errors.Is(err, redis.Nil))
}

func TestInvalidateFeatureFlag(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	require.NoError(t, c.SetFeatureFlag(ctx, "o", "f", true))
	require.NoError(t, c.InvalidateFeatureFlag(ctx, "o", "f"))
	_, err := mr.Get("feature_flag:o:f")
	require.Error(t, err)
}

func TestRBAC_miss_hit_invalidate(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	uid, oid := "user-1", "org-1"
	p, err := c.GetRBACPermissions(ctx, uid, oid)
	require.NoError(t, err)
	require.Nil(t, p)

	perms := &CachedRBACPermissions{Roles: []string{"a"}, Permissions: []string{"p"}}
	require.NoError(t, c.SetRBACPermissions(ctx, uid, oid, perms))
	got, err := c.GetRBACPermissions(ctx, uid, oid)
	require.NoError(t, err)
	require.Equal(t, perms.Roles, got.Roles)
	require.Equal(t, perms.Permissions, got.Permissions)

	require.NoError(t, c.InvalidateRBACPermissions(ctx, uid, oid))
	_, err = mr.Get("rbac:user-1:org-1")
	require.Error(t, err)
}

func TestGetRBACPermissions_invalidJSON(t *testing.T) {
	c, mr := testCache(t)
	ctx := context.Background()
	uid, oid := "u2", "o2"
	key := "rbac:" + uid + ":" + oid
	mr.Set(key, "not-json")
	_, err := c.GetRBACPermissions(ctx, uid, oid)
	require.Error(t, err)
	_, err = mr.Get(key)
	require.Error(t, err)
}

func TestGetRBACPermissions_redisError(t *testing.T) {
	c, mr := testCache(t)
	mr.Close()
	_, err := c.GetRBACPermissions(context.Background(), "u", "o")
	require.Error(t, err)
}

func TestSetRBACPermissions_nil(t *testing.T) {
	c, _ := testCache(t)
	require.NoError(t, c.SetRBACPermissions(context.Background(), "u", "o", nil))
}

func TestSetRBACPermissions_marshalError(t *testing.T) {
	c, _ := testCache(t)
	swapJSONMarshal(t, func(any) ([]byte, error) { return nil, errors.New("marshal") })
	perms := &CachedRBACPermissions{Roles: []string{"r"}}
	require.Error(t, c.SetRBACPermissions(context.Background(), "u", "o", perms))
}

func TestSetRBACPermissions_redisError(t *testing.T) {
	c, mr := testCache(t)
	mr.Close()
	perms := &CachedRBACPermissions{Roles: []string{"r"}}
	require.Error(t, c.SetRBACPermissions(context.Background(), "u", "o", perms))
}

func TestSession_miss_hit(t *testing.T) {
	c, _ := testCache(t)
	ctx := context.Background()
	key := "sess-key"
	b, err := c.GetSession(ctx, key)
	require.NoError(t, err)
	require.Nil(t, b)

	payload := []byte(`{"ok":true}`)
	require.NoError(t, c.SetSession(ctx, key, payload))
	got, err := c.GetSession(ctx, key)
	require.NoError(t, err)
	require.Equal(t, payload, got)
}

func TestGetSession_redisError(t *testing.T) {
	c, mr := testCache(t)
	mr.Close()
	_, err := c.GetSession(context.Background(), "k")
	require.Error(t, err)
}

func TestSetSession_redisError(t *testing.T) {
	c, mr := testCache(t)
	mr.Close()
	require.Error(t, c.SetSession(context.Background(), "k", []byte("x")))
}
