package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/redisclient"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	UserCacheKeyPrefix          = "user:"
	OrgMembershipCacheKeyPrefix = "org_membership:"
	UserCacheTTL                = 10 * time.Minute
	OrgMembershipCacheTTL       = 30 * time.Minute
	FeatureFlagCacheKeyPrefix   = "feature_flag:"
	FeatureFlagCacheTTL         = 10 * time.Minute
)

type Cache struct {
	client *redis.Client
}

type CacheRepository interface {
	GetUser(ctx context.Context, email string) (*types.User, error)
	SetUser(ctx context.Context, email string, user *types.User) error
	GetOrgMembership(ctx context.Context, userID, orgID string) (bool, error)
	SetOrgMembership(ctx context.Context, userID, orgID string, belongs bool) error
	GetFeatureFlag(ctx context.Context, orgID, featureName string) (bool, error)
	SetFeatureFlag(ctx context.Context, orgID, featureName string, enabled bool) error
	InvalidateFeatureFlag(ctx context.Context, orgID, featureName string) error
}

func NewCache(redisURL string) (*Cache, error) {
	client, err := redisclient.New(redisURL)
	if err != nil {
		return nil, err
	}
	return &Cache{client: client}, nil
}

func (c *Cache) GetUser(ctx context.Context, email string) (*types.User, error) {
	key := UserCacheKeyPrefix + email
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var user types.User
	if err := json.Unmarshal(data, &user); err != nil {
		_ = c.InvalidateUser(ctx, email)
		return nil, err
	}

	if user.ID == uuid.Nil {
		_ = c.InvalidateUser(ctx, email)
		return nil, nil
	}

	return &user, nil
}

func (c *Cache) SetUser(ctx context.Context, email string, user *types.User) error {
	key := UserCacheKeyPrefix + email

	userCopy := *user

	if userCopy.ID == uuid.Nil {
		return nil
	}

	data, err := json.Marshal(userCopy)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, UserCacheTTL).Err()
}

func (c *Cache) GetOrgMembership(ctx context.Context, userID, orgID string) (bool, error) {
	key := OrgMembershipCacheKeyPrefix + userID + ":" + orgID
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

func (c *Cache) SetOrgMembership(ctx context.Context, userID, orgID string, belongs bool) error {
	key := OrgMembershipCacheKeyPrefix + userID + ":" + orgID
	val := "false"
	if belongs {
		val = "true"
	}
	return c.client.Set(ctx, key, val, OrgMembershipCacheTTL).Err()
}

func (c *Cache) InvalidateUser(ctx context.Context, email string) error {
	key := UserCacheKeyPrefix + email
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) InvalidateOrgMembership(ctx context.Context, userID, orgID string) error {
	key := OrgMembershipCacheKeyPrefix + userID + ":" + orgID
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) GetFeatureFlag(ctx context.Context, orgID, featureName string) (bool, error) {
	key := fmt.Sprintf("%s:%s:%s", FeatureFlagCacheKeyPrefix, orgID, featureName)
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, redis.Nil
	}
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

func (c *Cache) SetFeatureFlag(ctx context.Context, orgID, featureName string, enabled bool) error {
	key := fmt.Sprintf("%s:%s:%s", FeatureFlagCacheKeyPrefix, orgID, featureName)
	val := "false"
	if enabled {
		val = "true"
	}
	return c.client.Set(ctx, key, val, FeatureFlagCacheTTL).Err()
}

func (c *Cache) InvalidateFeatureFlag(ctx context.Context, orgID, featureName string) error {
	key := fmt.Sprintf("%s:%s:%s", FeatureFlagCacheKeyPrefix, orgID, featureName)
	return c.client.Del(ctx, key).Err()
}
