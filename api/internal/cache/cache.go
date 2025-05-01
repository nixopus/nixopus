package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	UserCacheKeyPrefix          = "user:"
	OrgMembershipCacheKeyPrefix = "org_membership:"
	UserCacheTTL                = 10 * time.Minute
	OrgMembershipCacheTTL       = 30 * time.Minute
)

type Cache struct {
	client *redis.Client
}

type CacheRepository interface {
	GetUser(ctx context.Context, email string) (*types.User, error)
	SetUser(ctx context.Context, email string, user *types.User) error
	GetOrgMembership(ctx context.Context, userID, orgID string) (bool, error)
	SetOrgMembership(ctx context.Context, userID, orgID string, belongs bool) error
}

func NewCache(redisURL string) (*Cache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	return &Cache{client: client}, nil
}

func (c *Cache) GetUser(ctx context.Context, email string) (*types.User, error) {
	key := UserCacheKeyPrefix + email
	data, err := c.client.Get(key).Bytes()
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

	return c.client.Set(key, data, UserCacheTTL).Err()
}

func (c *Cache) GetOrgMembership(ctx context.Context, userID, orgID string) (bool, error) {
	key := OrgMembershipCacheKeyPrefix + userID + ":" + orgID
	val, err := c.client.Get(key).Result()
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
	return c.client.Set(key, val, OrgMembershipCacheTTL).Err()
}

func (c *Cache) InvalidateUser(ctx context.Context, email string) error {
	key := UserCacheKeyPrefix + email
	return c.client.Del(key).Err()
}

func (c *Cache) InvalidateOrgMembership(ctx context.Context, userID, orgID string) error {
	key := OrgMembershipCacheKeyPrefix + userID + ":" + orgID
	return c.client.Del(key).Err()
}
