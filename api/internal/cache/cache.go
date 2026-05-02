package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/redisclient"
	"github.com/nixopus/nixopus/api/internal/types"
)

// jsonMarshal is json.Marshal by default; tests swap it to cover marshal error paths.
var jsonMarshal = json.Marshal

const (
	UserCacheKeyPrefix          = "user:"
	UserByIDCacheKeyPrefix      = "user_id:"
	OrgMembershipCacheKeyPrefix = "org_membership:"
	UserCacheTTL                = 10 * time.Minute
	OrgMembershipCacheTTL       = 30 * time.Minute
	FeatureFlagCacheKeyPrefix   = "feature_flag:"
	FeatureFlagCacheTTL         = 10 * time.Minute
	RBACCacheKeyPrefix          = "rbac:"
	RBACCacheTTL                = 5 * time.Minute
	SessionCacheKeyPrefix       = "session:"
	SessionCacheTTL             = 5 * time.Minute

	AdminRegisteredCacheKey = "auth:admin_registered"
	// AdminRegisteredTrueTTL: once an admin exists the value is permanent; cache aggressively.
	AdminRegisteredTrueTTL = 24 * time.Hour
	// AdminRegisteredFalseTTL: before first signup, re-check soon so signup is detected quickly.
	AdminRegisteredFalseTTL = 30 * time.Second

	ExtensionByIDCacheKeyPrefix    = "ext:id:"
	ExtensionByExtIDCacheKeyPrefix = "ext:eid:"
	ExtensionCategoriesCacheKey    = "ext:categories"
	ExtensionCacheTTL              = 15 * time.Minute
	ExtensionCategoriesCacheTTL    = 1 * time.Hour
)

type Cache struct {
	client *redis.Client
}

// CachedRBACPermissions stores user permissions for an organization
type CachedRBACPermissions struct {
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type CacheRepository interface {
	GetUser(ctx context.Context, email string) (*types.User, error)
	SetUser(ctx context.Context, email string, user *types.User) error
	GetOrgMembership(ctx context.Context, userID, orgID string) (bool, error)
	SetOrgMembership(ctx context.Context, userID, orgID string, belongs bool) error
	GetFeatureFlag(ctx context.Context, orgID, featureName string) (bool, error)
	SetFeatureFlag(ctx context.Context, orgID, featureName string, enabled bool) error
	InvalidateFeatureFlag(ctx context.Context, orgID, featureName string) error
	GetRBACPermissions(ctx context.Context, userID, orgID string) (*CachedRBACPermissions, error)
	SetRBACPermissions(ctx context.Context, userID, orgID string, perms *CachedRBACPermissions) error
	InvalidateRBACPermissions(ctx context.Context, userID, orgID string) error
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

	data, err := jsonMarshal(userCopy)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, UserCacheTTL).Err()
}

// GetUserByID retrieves a cached user by their UUID.
func (c *Cache) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	key := UserByIDCacheKeyPrefix + userID
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var user types.User
	if err := json.Unmarshal(data, &user); err != nil {
		_ = c.client.Del(ctx, key).Err()
		return nil, err
	}

	if user.ID == uuid.Nil {
		_ = c.client.Del(ctx, key).Err()
		return nil, nil
	}

	return &user, nil
}

// SetUserByID caches a user record by their UUID.
func (c *Cache) SetUserByID(ctx context.Context, userID string, user *types.User) error {
	if user == nil || user.ID == uuid.Nil {
		return nil
	}
	key := UserByIDCacheKeyPrefix + userID
	data, err := jsonMarshal(user)
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

// InvalidateUserByID removes a cached user record keyed by UUID.
// This must be called whenever user fields (name, avatar, is_onboarded,
// provision_status, etc.) are modified so the auth middleware re-reads
// the row from the database on the next request.
func (c *Cache) InvalidateUserByID(ctx context.Context, userID string) error {
	key := UserByIDCacheKeyPrefix + userID
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

// GetRBACPermissions retrieves cached RBAC permissions for a user in an organization.
// Returns nil if not found in cache (cache miss).
func (c *Cache) GetRBACPermissions(ctx context.Context, userID, orgID string) (*CachedRBACPermissions, error) {
	key := fmt.Sprintf("%s%s:%s", RBACCacheKeyPrefix, userID, orgID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var perms CachedRBACPermissions
	if err := json.Unmarshal(data, &perms); err != nil {
		_ = c.InvalidateRBACPermissions(ctx, userID, orgID)
		return nil, err
	}

	return &perms, nil
}

// SetRBACPermissions caches RBAC permissions for a user in an organization.
func (c *Cache) SetRBACPermissions(ctx context.Context, userID, orgID string, perms *CachedRBACPermissions) error {
	if perms == nil {
		return nil
	}

	key := fmt.Sprintf("%s%s:%s", RBACCacheKeyPrefix, userID, orgID)
	data, err := jsonMarshal(perms)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, RBACCacheTTL).Err()
}

// InvalidateRBACPermissions removes cached RBAC permissions for a user in an organization.
func (c *Cache) InvalidateRBACPermissions(ctx context.Context, userID, orgID string) error {
	key := fmt.Sprintf("%s%s:%s", RBACCacheKeyPrefix, userID, orgID)
	return c.client.Del(ctx, key).Err()
}

// GetSession retrieves a cached session verification result.
// Returns nil on cache miss.
func (c *Cache) GetSession(ctx context.Context, cacheKey string) ([]byte, error) {
	key := SessionCacheKeyPrefix + cacheKey
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SetSession caches a session verification result.
func (c *Cache) SetSession(ctx context.Context, cacheKey string, data []byte) error {
	key := SessionCacheKeyPrefix + cacheKey
	return c.client.Set(ctx, key, data, SessionCacheTTL).Err()
}

// GetAdminRegistered returns the cached value and whether a cache hit occurred.
func (c *Cache) GetAdminRegistered(ctx context.Context) (registered bool, hit bool, err error) {
	val, err := c.client.Get(ctx, AdminRegisteredCacheKey).Result()
	if err == redis.Nil {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return val == "true", true, nil
}

// SetAdminRegistered caches the result with a TTL that depends on the value.
// true  -> long TTL (state effectively permanent once an admin exists)
// false -> short TTL (re-check soon so first signup is detected quickly)
func (c *Cache) SetAdminRegistered(ctx context.Context, registered bool) error {
	val := "false"
	ttl := AdminRegisteredFalseTTL
	if registered {
		val = "true"
		ttl = AdminRegisteredTrueTTL
	}
	return c.client.Set(ctx, AdminRegisteredCacheKey, val, ttl).Err()
}

// InvalidateAdminRegistered removes the cached admin registration status.
func (c *Cache) InvalidateAdminRegistered(ctx context.Context) error {
	return c.client.Del(ctx, AdminRegisteredCacheKey).Err()
}

// --- Extensions ---

func (c *Cache) GetExtension(ctx context.Context, id string) (*types.Extension, error) {
	data, err := c.client.Get(ctx, ExtensionByIDCacheKeyPrefix+id).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ext types.Extension
	if err := json.Unmarshal(data, &ext); err != nil {
		_ = c.client.Del(ctx, ExtensionByIDCacheKeyPrefix+id).Err()
		return nil, err
	}
	return &ext, nil
}

func (c *Cache) GetExtensionByExtID(ctx context.Context, extensionID string) (*types.Extension, error) {
	data, err := c.client.Get(ctx, ExtensionByExtIDCacheKeyPrefix+extensionID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ext types.Extension
	if err := json.Unmarshal(data, &ext); err != nil {
		_ = c.client.Del(ctx, ExtensionByExtIDCacheKeyPrefix+extensionID).Err()
		return nil, err
	}
	return &ext, nil
}

func (c *Cache) GetExtensionCategories(ctx context.Context) ([]types.ExtensionCategory, error) {
	data, err := c.client.Get(ctx, ExtensionCategoriesCacheKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cats []types.ExtensionCategory
	if err := json.Unmarshal(data, &cats); err != nil {
		_ = c.client.Del(ctx, ExtensionCategoriesCacheKey).Err()
		return nil, err
	}
	return cats, nil
}

func (c *Cache) SetExtension(ctx context.Context, ext *types.Extension) error {
	data, err := jsonMarshal(ext)
	if err != nil {
		return err
	}
	pipe := c.client.Pipeline()
	pipe.Set(ctx, ExtensionByIDCacheKeyPrefix+ext.ID.String(), data, ExtensionCacheTTL)
	pipe.Set(ctx, ExtensionByExtIDCacheKeyPrefix+ext.ExtensionID, data, ExtensionCacheTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *Cache) SetExtensionCategories(ctx context.Context, cats []types.ExtensionCategory) error {
	data, err := jsonMarshal(cats)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, ExtensionCategoriesCacheKey, data, ExtensionCategoriesCacheTTL).Err()
}

// InvalidateExtension removes both ID-based and ExtensionID-based cache entries.
// Both keys must be cleared because the same entity is reachable via two lookup paths.
func (c *Cache) InvalidateExtension(ctx context.Context, id string, extensionID string) error {
	keys := make([]string, 0, 2)
	if id != "" {
		keys = append(keys, ExtensionByIDCacheKeyPrefix+id)
	}
	if extensionID != "" {
		keys = append(keys, ExtensionByExtIDCacheKeyPrefix+extensionID)
	}
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// InvalidateExtensionCategories removes the cached categories list.
// Must be called whenever an extension is created, deleted, or changes category.
func (c *Cache) InvalidateExtensionCategories(ctx context.Context) error {
	return c.client.Del(ctx, ExtensionCategoriesCacheKey).Err()
}
