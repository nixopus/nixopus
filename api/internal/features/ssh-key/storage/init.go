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

type SSHKeyRepository interface {
	CreateSSHKey(sshKey *types.SSHKey) error
	GetSSHKeyByID(id uuid.UUID) (*types.SSHKey, error)
	GetSSHKeysByOrganizationID(organizationID uuid.UUID) ([]*types.SSHKey, error)
	UpdateSSHKey(sshKey *types.SSHKey) error
	DeleteSSHKey(id uuid.UUID) error
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) SSHKeyRepository
}

func NewSSHKeyStorage(db *bun.DB, ctx context.Context) *SSHKeyStorage {
	return &SSHKeyStorage{
		DB:  db,
		Ctx: ctx,
	}
}

func (s *SSHKeyStorage) WithTx(tx bun.Tx) SSHKeyRepository {
	return &SSHKeyStorage{
		DB:  s.DB,
		Ctx: s.Ctx,
		tx:  &tx,
	}
}

func (s *SSHKeyStorage) BeginTx() (bun.Tx, error) {
	return s.DB.BeginTx(s.Ctx, nil)
}

func (s *SSHKeyStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

// CreateSSHKey creates a new SSH key in the database
func (s *SSHKeyStorage) CreateSSHKey(sshKey *types.SSHKey) error {
	_, err := s.getDB().NewInsert().Model(sshKey).Exec(s.Ctx)
	return err
}

// GetSSHKeyByID retrieves an SSH key by ID
func (s *SSHKeyStorage) GetSSHKeyByID(id uuid.UUID) (*types.SSHKey, error) {
	sshKey := &types.SSHKey{}
	err := s.getDB().NewSelect().
		Model(sshKey).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return sshKey, nil
}

// GetSSHKeysByOrganizationID retrieves all SSH keys for an organization
func (s *SSHKeyStorage) GetSSHKeysByOrganizationID(organizationID uuid.UUID) ([]*types.SSHKey, error) {
	var sshKeys []*types.SSHKey
	err := s.getDB().NewSelect().
		Model(&sshKeys).
		Where("organization_id = ?", organizationID).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return sshKeys, nil
}

// UpdateSSHKey updates an SSH key
func (s *SSHKeyStorage) UpdateSSHKey(sshKey *types.SSHKey) error {
	_, err := s.getDB().NewUpdate().
		Model(sshKey).
		Where("id = ?", sshKey.ID).
		Exec(s.Ctx)
	return err
}

// DeleteSSHKey performs a soft delete on an SSH key
func (s *SSHKeyStorage) DeleteSSHKey(id uuid.UUID) error {
	_, err := s.getDB().NewUpdate().
		Model((*types.SSHKey)(nil)).
		Set("deleted_at = NOW()").
		Set("updated_at = NOW()").
		Where("id = ?", id).
		Exec(s.Ctx)
	return err
}
