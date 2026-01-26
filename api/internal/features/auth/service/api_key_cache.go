package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// APIKeyCache provides caching for API key validation to improve performance
type APIKeyCache struct {
	cache      map[string]*cacheEntry
	mu         sync.RWMutex
	ttl        time.Duration
	logger     logger.Logger
	maxEntries int
}

type cacheEntry struct {
	apiKey    *shared_types.APIKey
	expiresAt time.Time
}

// NewAPIKeyCache creates a new API key cache
func NewAPIKeyCache(ttl time.Duration, maxEntries int, logger logger.Logger) *APIKeyCache {
	cache := &APIKeyCache{
		cache:      make(map[string]*cacheEntry),
		ttl:        ttl,
		logger:     logger,
		maxEntries: maxEntries,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves an API key from cache if valid
func (c *APIKeyCache) Get(keyHash string) (*shared_types.APIKey, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[keyHash]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	// Check if API key is still valid
	if !entry.apiKey.IsValid() {
		return nil, false
	}

	return entry.apiKey, true
}

// Set stores an API key in cache
func (c *APIKeyCache) Set(keyHash string, apiKey *shared_types.APIKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if cache is full
	if len(c.cache) >= c.maxEntries {
		c.evictOldest()
	}

	c.cache[keyHash] = &cacheEntry{
		apiKey:    apiKey,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes an API key from cache (e.g., when revoked)
func (c *APIKeyCache) Invalidate(keyHash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, keyHash)
}

// evictOldest removes the oldest entry from cache
func (c *APIKeyCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range c.cache {
		if first || entry.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expiresAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

// cleanup periodically removes expired entries
func (c *APIKeyCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.After(entry.expiresAt) || !entry.apiKey.IsValid() {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

// InvalidateByUserID removes all API keys for a user from cache
func (c *APIKeyCache) InvalidateByUserID(userID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.cache {
		if entry.apiKey.UserID == userID {
			delete(c.cache, key)
		}
	}
}

// InvalidateByKeyID removes a specific API key by ID from cache
func (c *APIKeyCache) InvalidateByKeyID(keyID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.cache {
		if entry.apiKey.ID == keyID {
			delete(c.cache, key)
		}
	}
}
