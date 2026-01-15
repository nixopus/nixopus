package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type APIKeyStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

func (s *APIKeyStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

// CreateAPIKey creates a new API key in the database
func (s *APIKeyStorage) CreateAPIKey(apiKey *types.APIKey) error {
	_, err := s.getDB().NewInsert().Model(apiKey).Exec(s.Ctx)
	return err
}

// FindAPIKeyByHash finds an API key by its hash
func (s *APIKeyStorage) FindAPIKeyByHash(keyHash string) (*types.APIKey, error) {
	apiKey := &types.APIKey{}
	err := s.getDB().NewSelect().
		Model(apiKey).
		Where("key_hash = ?", keyHash).
		Relation("User").
		Relation("Organization").
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}

// FindAPIKeysByUserID finds all API keys for a user
func (s *APIKeyStorage) FindAPIKeysByUserID(userID uuid.UUID) ([]*types.APIKey, error) {
	var apiKeys []*types.APIKey
	err := s.getDB().NewSelect().
		Model(&apiKeys).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Order("created_at DESC").
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return apiKeys, nil
}

// FindAPIKeysByOrganizationID finds all API keys for an organization
func (s *APIKeyStorage) FindAPIKeysByOrganizationID(organizationID uuid.UUID) ([]*types.APIKey, error) {
	var apiKeys []*types.APIKey
	err := s.getDB().NewSelect().
		Model(&apiKeys).
		Where("organization_id = ?", organizationID).
		Where("revoked_at IS NULL").
		Order("created_at DESC").
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return apiKeys, nil
}

// RevokeAPIKey revokes an API key by setting revoked_at
func (s *APIKeyStorage) RevokeAPIKey(keyID uuid.UUID) error {
	now := time.Now()
	_, err := s.getDB().NewUpdate().
		Model(&types.APIKey{}).
		Set("revoked_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", keyID).
		Exec(s.Ctx)
	return err
}

// UpdateLastUsed updates the last_used_at timestamp for an API key
func (s *APIKeyStorage) UpdateLastUsed(keyID uuid.UUID) error {
	now := time.Now()
	_, err := s.getDB().NewUpdate().
		Model(&types.APIKey{}).
		Set("last_used_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", keyID).
		Exec(s.Ctx)
	return err
}

// DeleteAPIKey permanently deletes an API key
func (s *APIKeyStorage) DeleteAPIKey(keyID uuid.UUID) error {
	_, err := s.getDB().NewDelete().
		Model(&types.APIKey{}).
		Where("id = ?", keyID).
		Exec(s.Ctx)
	return err
}
