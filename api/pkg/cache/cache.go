// Package cache provides public access to cache types from the internal package.
package cache

import (
	internalCache "github.com/raghavyuva/nixopus-api/internal/cache"
)

// Cache is a type alias for the internal cache.Cache type.
// This allows other modules to reference the Cache type without importing internal packages.
type Cache = internalCache.Cache

// CachedRBACPermissions is a type alias for the internal cache.CachedRBACPermissions type.
type CachedRBACPermissions = internalCache.CachedRBACPermissions

// NewCache creates a new Cache instance.
func NewCache(redisURL string) (*Cache, error) {
	return internalCache.NewCache(redisURL)
}
