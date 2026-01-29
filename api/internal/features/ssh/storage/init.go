package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type SSHKeyStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

func (s *SSHKeyStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

// SSHKeyRepository defines the interface for SSH key storage operations.
// This enables mocking in tests.
type SSHKeyRepository interface {
	GetActiveSSHKeyByOrganizationID(orgID uuid.UUID) (*types.SSHKey, error)
	GetSSHKeyByID(keyID uuid.UUID) (*types.SSHKey, error)
	ListSSHKeysByOrganizationID(orgID uuid.UUID) ([]*types.SSHKey, error)
}

// GetActiveSSHKeyByOrganizationID retrieves the most recent active SSH key for an organization.
// Returns sql.ErrNoRows if no active key is found.
func (s *SSHKeyStorage) GetActiveSSHKeyByOrganizationID(orgID uuid.UUID) (*types.SSHKey, error) {
	var sshKey types.SSHKey
	err := s.getDB().NewSelect().
		Model(&sshKey).
		Where("organization_id = ?", orgID).
		Where("is_active = ?", true).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(1).
		Scan(s.Ctx)

	if err != nil {
		return nil, err
	}
	return &sshKey, nil
}

// GetSSHKeyByID retrieves an SSH key by its ID.
// Returns sql.ErrNoRows if the key is not found.
func (s *SSHKeyStorage) GetSSHKeyByID(keyID uuid.UUID) (*types.SSHKey, error) {
	var sshKey types.SSHKey
	err := s.getDB().NewSelect().
		Model(&sshKey).
		Where("id = ?", keyID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)

	if err != nil {
		return nil, err
	}
	return &sshKey, nil
}

// ListSSHKeysByOrganizationID retrieves all SSH keys (including inactive) for an organization.
// Excludes soft-deleted keys.
func (s *SSHKeyStorage) ListSSHKeysByOrganizationID(orgID uuid.UUID) ([]*types.SSHKey, error) {
	var sshKeys []*types.SSHKey
	err := s.getDB().NewSelect().
		Model(&sshKeys).
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Scan(s.Ctx)

	if err != nil {
		return nil, err
	}
	return sshKeys, nil
}
