package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type APIKeyService struct {
	storage storage.APIKeyStorage
	logger  logger.Logger
	ctx     interface{}
	cache   *APIKeyCache
}

func NewAPIKeyService(apiKeyStorage storage.APIKeyStorage, logger logger.Logger) *APIKeyService {
	// Create cache with 5 minute TTL and max 1000 entries
	cache := NewAPIKeyCache(5*time.Minute, 1000, logger)
	return &APIKeyService{
		storage: apiKeyStorage,
		logger:  logger,
		cache:   cache,
	}
}

// GenerateAPIKey generates a new API key and returns both the full key and the APIKey model
// The full key is only returned once during creation
func (s *APIKeyService) GenerateAPIKey(userID, organizationID uuid.UUID, name string, expiresInDays *int) (string, *shared_types.APIKey, error) {
	// Generate a secure random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Encode to base64 for the actual key
	key := base64.URLEncoding.EncodeToString(keyBytes)

	// Create prefix (first 8 characters) for display purposes
	prefix := key[:8]

	// Format the key with prefix for user display: nixopus_<prefix>_<rest>
	formattedKey := fmt.Sprintf("nixopus_%s_%s", prefix, key[8:])

	// Hash the full formatted key for storage
	hash := sha256.Sum256([]byte(formattedKey))
	keyHash := hex.EncodeToString(hash[:])

	// Set expiration if provided
	var expiresAt *time.Time
	if expiresInDays != nil && *expiresInDays > 0 {
		exp := time.Now().Add(time.Duration(*expiresInDays) * 24 * time.Hour)
		expiresAt = &exp
	}

	apiKey := &shared_types.APIKey{
		ID:             uuid.New(),
		UserID:         userID,
		OrganizationID: organizationID,
		Name:           name,
		KeyHash:        keyHash,
		Prefix:         prefix,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.storage.CreateAPIKey(apiKey); err != nil {
		return "", nil, fmt.Errorf("failed to create API key: %w", err)
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Created API key %s for user %s", apiKey.ID, userID), "")

	return formattedKey, apiKey, nil
}

// VerifyAPIKey verifies an API key and returns the associated APIKey model
// Uses caching for improved performance
func (s *APIKeyService) VerifyAPIKey(key string) (*shared_types.APIKey, error) {
	// Parse the formatted key: nixopus_<prefix>_<rest>
	if len(key) < 9 || key[:8] != "nixopus_" {
		return nil, fmt.Errorf("invalid API key format")
	}

	// Hash the full provided key
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	// Try cache first
	if cached, found := s.cache.Get(keyHash); found {
		// Update last used timestamp asynchronously to avoid blocking
		go func() {
			if err := s.storage.UpdateLastUsed(cached.ID); err != nil {
				s.logger.Log(logger.Warning, fmt.Sprintf("Failed to update last used for API key %s: %v", cached.ID, err), "")
			}
		}()
		return cached, nil
	}

	// Cache miss - query database
	apiKey, err := s.storage.FindAPIKeyByHash(keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check if key is valid
	if !apiKey.IsValid() {
		return nil, fmt.Errorf("API key is revoked or expired")
	}

	// Store in cache
	s.cache.Set(keyHash, apiKey)

	// Update last used timestamp
	if err := s.storage.UpdateLastUsed(apiKey.ID); err != nil {
		s.logger.Log(logger.Warning, fmt.Sprintf("Failed to update last used for API key %s: %v", apiKey.ID, err), "")
	}

	return apiKey, nil
}

// ListAPIKeys lists all API keys for a user
func (s *APIKeyService) ListAPIKeys(userID uuid.UUID) ([]*shared_types.APIKey, error) {
	return s.storage.FindAPIKeysByUserID(userID)
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(keyID uuid.UUID) error {
	if err := s.storage.RevokeAPIKey(keyID); err != nil {
		return err
	}
	// Invalidate cache
	s.cache.InvalidateByKeyID(keyID)
	return nil
}

// DeleteAPIKey permanently deletes an API key
func (s *APIKeyService) DeleteAPIKey(keyID uuid.UUID) error {
	return s.storage.DeleteAPIKey(keyID)
}
