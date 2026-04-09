package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	api_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type RegistrationStorage struct {
	db  *bun.DB
	ctx context.Context
}

func NewRegistrationStorage(db *bun.DB, ctx context.Context) *RegistrationStorage {
	return &RegistrationStorage{db: db, ctx: ctx}
}

func (s *RegistrationStorage) CountUserOwnedMachines(orgID uuid.UUID) (int, error) {
	count, err := s.db.NewSelect().
		TableExpr("user_provision_details").
		Where("organization_id = ?", orgID).
		Where("type = 'user_owned'").
		Count(s.ctx)
	return count, err
}

func (s *RegistrationStorage) HostPortExists(orgID uuid.UUID, host string, port int) (bool, error) {
	exists, err := s.db.NewSelect().
		Model((*api_types.SSHKey)(nil)).
		Where("organization_id = ?", orgID).
		Where("host = ?", host).
		Where("port = ?", port).
		Where("deleted_at IS NULL").
		Exists(s.ctx)
	return exists, err
}

func (s *RegistrationStorage) InsertSSHKey(key *api_types.SSHKey) error {
	_, err := s.db.NewInsert().Model(key).Exec(s.ctx)
	return err
}

func (s *RegistrationStorage) InsertProvisionDetails(userID, orgID, sshKeyID uuid.UUID, provisionType, step string) error {
	_, err := s.db.NewInsert().
		TableExpr("user_provision_details").
		Value("id", "uuid_generate_v4()").
		Value("user_id", "?", userID).
		Value("organization_id", "?", orgID).
		Value("ssh_key_id", "?", sshKeyID).
		Value("type", "?", provisionType).
		Value("step", "?", step).
		Value("created_at", "NOW()").
		Value("updated_at", "NOW()").
		Exec(s.ctx)
	return err
}

func (s *RegistrationStorage) GetSSHKeyByID(id, orgID uuid.UUID) (*api_types.SSHKey, error) {
	var key api_types.SSHKey
	err := s.db.NewSelect().
		Model(&key).
		Where("id = ?", id).
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Scan(s.ctx)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *RegistrationStorage) GetSSHKeyStatus(id, orgID uuid.UUID) (bool, *time.Time, error) {
	var key api_types.SSHKey
	err := s.db.NewSelect().
		Model(&key).
		Column("is_active", "last_used_at").
		Where("id = ?", id).
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Scan(s.ctx)
	if err != nil {
		return false, nil, err
	}
	return key.IsActive, key.LastUsedAt, nil
}

func (s *RegistrationStorage) HasActiveAppServers(sshKeyID uuid.UUID) (bool, error) {
	exists, err := s.db.NewSelect().
		TableExpr("application_servers").
		Where("server_id = ?", sshKeyID).
		Exists(s.ctx)
	return exists, err
}

func (s *RegistrationStorage) SoftDeleteSSHKey(id uuid.UUID) error {
	_, err := s.db.NewUpdate().
		Model((*api_types.SSHKey)(nil)).
		Set("deleted_at = ?", time.Now()).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(s.ctx)
	return err
}

func (s *RegistrationStorage) GetStaleBYOSMachines(orgID uuid.UUID, cutoff time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := s.db.NewSelect().
		TableExpr("ssh_keys AS sk").
		ColumnExpr("sk.id").
		Join("JOIN user_provision_details AS upd ON upd.ssh_key_id = sk.id").
		Where("upd.organization_id = ?", orgID).
		Where("upd.type = 'user_owned'").
		Where("sk.is_active = false").
		Where("sk.created_at < ?", cutoff).
		Where("sk.deleted_at IS NULL").
		Scan(s.ctx, &ids)
	return ids, err
}

func (s *RegistrationStorage) GetActiveUserOwnedMachines(orgID uuid.UUID) ([]api_types.SSHKey, error) {
	var keys []api_types.SSHKey
	err := s.db.NewSelect().
		Model(&keys).
		Join("JOIN user_provision_details AS upd ON upd.ssh_key_id = ssh_key.id").
		Where("upd.organization_id = ?", orgID).
		Where("upd.type = 'user_owned'").
		Where("ssh_key.is_active = true").
		Where("ssh_key.deleted_at IS NULL").
		Scan(s.ctx)
	return keys, err
}

func (s *RegistrationStorage) GetProvisionDetailsBySSHKeyID(sshKeyID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.NewSelect().
		TableExpr("user_provision_details").
		Column("id").
		Where("ssh_key_id = ?", sshKeyID).
		Scan(s.ctx, &id)
	return id, err
}
