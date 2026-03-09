package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nixopus/nixopus/api/internal/redisclient"
	"github.com/nixopus/nixopus/api/internal/types"
)

const (
	extensionByIDPrefix    = "ext:id:"
	extensionByExtIDPrefix = "ext:eid:"
	categoriesKey          = "ext:categories"

	extensionCacheTTL  = 15 * time.Minute
	categoriesCacheTTL = 1 * time.Hour
)

type ExtensionCache struct {
	client *redis.Client
}

func NewExtensionCache(redisURL string) (*ExtensionCache, error) {
	client, err := redisclient.New(redisURL)
	if err != nil {
		return nil, err
	}
	return &ExtensionCache{client: client}, nil
}

// --- Get methods ---

func (c *ExtensionCache) GetExtension(ctx context.Context, id string) (*types.Extension, error) {
	data, err := c.client.Get(ctx, extensionByIDPrefix+id).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ext types.Extension
	if err := json.Unmarshal(data, &ext); err != nil {
		_ = c.client.Del(ctx, extensionByIDPrefix+id).Err()
		return nil, err
	}
	return &ext, nil
}

func (c *ExtensionCache) GetExtensionByExtID(ctx context.Context, extensionID string) (*types.Extension, error) {
	data, err := c.client.Get(ctx, extensionByExtIDPrefix+extensionID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ext types.Extension
	if err := json.Unmarshal(data, &ext); err != nil {
		_ = c.client.Del(ctx, extensionByExtIDPrefix+extensionID).Err()
		return nil, err
	}
	return &ext, nil
}

func (c *ExtensionCache) GetCategories(ctx context.Context) ([]types.ExtensionCategory, error) {
	data, err := c.client.Get(ctx, categoriesKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cats []types.ExtensionCategory
	if err := json.Unmarshal(data, &cats); err != nil {
		_ = c.client.Del(ctx, categoriesKey).Err()
		return nil, err
	}
	return cats, nil
}

// --- Set methods ---

func (c *ExtensionCache) SetExtension(ctx context.Context, ext *types.Extension) error {
	data, err := json.Marshal(ext)
	if err != nil {
		return err
	}
	pipe := c.client.Pipeline()
	pipe.Set(ctx, extensionByIDPrefix+ext.ID.String(), data, extensionCacheTTL)
	pipe.Set(ctx, extensionByExtIDPrefix+ext.ExtensionID, data, extensionCacheTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *ExtensionCache) SetCategories(ctx context.Context, cats []types.ExtensionCategory) error {
	data, err := json.Marshal(cats)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, categoriesKey, data, categoriesCacheTTL).Err()
}

// --- Invalidation methods ---

// InvalidateExtension removes both ID-based and ExtensionID-based cache entries.
// Both keys must be cleared because the same entity is reachable via two lookup paths.
func (c *ExtensionCache) InvalidateExtension(ctx context.Context, id string, extensionID string) error {
	keys := make([]string, 0, 2)
	if id != "" {
		keys = append(keys, extensionByIDPrefix+id)
	}
	if extensionID != "" {
		keys = append(keys, extensionByExtIDPrefix+extensionID)
	}
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// InvalidateCategories removes the cached categories list.
// Must be called whenever an extension is created, deleted, or changes category.
func (c *ExtensionCache) InvalidateCategories(ctx context.Context) error {
	return c.client.Del(ctx, categoriesKey).Err()
}
