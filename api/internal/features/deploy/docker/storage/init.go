package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type DockerConfigStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

func (s *DockerConfigStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

// DockerConfigRepository defines the interface for Docker config storage operations
// This enables mocking in tests.
type DockerConfigRepository interface {
	GetActiveDockerConfigByOrganizationID(orgID uuid.UUID) (*types.DockerConfig, error)
	GetDockerConfigByID(configID uuid.UUID) (*types.DockerConfig, error)
	ListDockerConfigsByOrganizationID(orgID uuid.UUID) ([]*types.DockerConfig, error)
	CreateDockerConfig(config *types.DockerConfig) error
	UpdateDockerConfig(configID uuid.UUID, updates map[string]interface{}) error
	DeleteDockerConfig(configID uuid.UUID) error
}

// GetActiveDockerConfigByOrganizationID retrieves the most recent active Docker config for an organization.
// Returns sql.ErrNoRows if no active config is found.
func (s *DockerConfigStorage) GetActiveDockerConfigByOrganizationID(orgID uuid.UUID) (*types.DockerConfig, error) {
	var dockerConfig types.DockerConfig
	err := s.getDB().NewSelect().
		Model(&dockerConfig).
		Where("organization_id = ?", orgID).
		Where("is_active = ?", true).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(1).
		Scan(s.Ctx)

	if err != nil {
		return nil, err
	}
	return &dockerConfig, nil
}

// GetDockerConfigByID retrieves a Docker config by its ID.
// Returns sql.ErrNoRows if the config is not found.
func (s *DockerConfigStorage) GetDockerConfigByID(configID uuid.UUID) (*types.DockerConfig, error) {
	var dockerConfig types.DockerConfig
	err := s.getDB().NewSelect().
		Model(&dockerConfig).
		Where("id = ?", configID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)

	if err != nil {
		return nil, err
	}
	return &dockerConfig, nil
}

// ListDockerConfigsByOrganizationID retrieves all Docker configs (including inactive) for an organization.
// Excludes soft-deleted configs.
func (s *DockerConfigStorage) ListDockerConfigsByOrganizationID(orgID uuid.UUID) ([]*types.DockerConfig, error) {
	var configs []*types.DockerConfig
	err := s.getDB().NewSelect().
		Model(&configs).
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Scan(s.Ctx)

	if err != nil {
		return nil, err
	}
	return configs, nil
}

// CreateDockerConfig creates a new Docker configuration
func (s *DockerConfigStorage) CreateDockerConfig(config *types.DockerConfig) error {
	_, err := s.getDB().NewInsert().Model(config).Exec(s.Ctx)
	return err
}

// UpdateDockerConfig updates a Docker configuration
func (s *DockerConfigStorage) UpdateDockerConfig(configID uuid.UUID, updates map[string]interface{}) error {
	query := s.getDB().NewUpdate().
		Model((*types.DockerConfig)(nil))

	for key, value := range updates {
		if key == "updated_at" {
			continue
		}
		query = query.Set("? = ?", bun.Ident(key), value)
	}

	query = query.Set("updated_at = NOW()")

	_, err := query.
		Where("id = ?", configID).
		Where("deleted_at IS NULL").
		Exec(s.Ctx)
	return err
}

// DeleteDockerConfig performs a soft delete
func (s *DockerConfigStorage) DeleteDockerConfig(configID uuid.UUID) error {
	_, err := s.getDB().NewUpdate().
		Model((*types.DockerConfig)(nil)).
		Set("deleted_at = NOW()").
		Set("updated_at = NOW()").
		Where("id = ?", configID).
		Where("deleted_at IS NULL").
		Exec(s.Ctx)
	return err
}
