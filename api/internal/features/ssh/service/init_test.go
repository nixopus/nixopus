package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh/service"
	"github.com/nixopus/nixopus/api/internal/features/ssh/storage"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ip(i int) *int { return &i }

func sp(s string) *string { return &s }

type mockSSHRepo struct {
	getDefaultFn func(uuid.UUID) (*types.SSHKey, error)
	getActiveFn  func(uuid.UUID) (*types.SSHKey, error)
	getByIDFn    func(uuid.UUID) (*types.SSHKey, error)
	listFn       func(uuid.UUID) ([]*types.SSHKey, error)
	promoteFn    func(uuid.UUID) error

	promoteCalls []uuid.UUID
}

func (m *mockSSHRepo) GetActiveSSHKeyByOrganizationID(orgID uuid.UUID) (*types.SSHKey, error) {
	if m.getActiveFn != nil {
		return m.getActiveFn(orgID)
	}
	return nil, sql.ErrNoRows
}

func (m *mockSSHRepo) GetDefaultSSHKeyByOrganizationID(orgID uuid.UUID) (*types.SSHKey, error) {
	if m.getDefaultFn != nil {
		return m.getDefaultFn(orgID)
	}
	return nil, sql.ErrNoRows
}

func (m *mockSSHRepo) GetSSHKeyByID(keyID uuid.UUID) (*types.SSHKey, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(keyID)
	}
	return nil, sql.ErrNoRows
}

func (m *mockSSHRepo) ListSSHKeysByOrganizationID(orgID uuid.UUID) ([]*types.SSHKey, error) {
	if m.listFn != nil {
		return m.listFn(orgID)
	}
	return nil, nil
}

func (m *mockSSHRepo) PromoteToDefault(keyID uuid.UUID) error {
	m.promoteCalls = append(m.promoteCalls, keyID)
	if m.promoteFn != nil {
		return m.promoteFn(keyID)
	}
	return nil
}

var _ storage.SSHKeyRepository = (*mockSSHRepo)(nil)

func TestGetSSHConfigForOrganization_defaultFound(t *testing.T) {
	org := uuid.New()
	host := "h.example"
	user := "ubuntu"
	priv := "-----BEGIN RSA PRIVATE KEY-----\nABC\n-----END RSA PRIVATE KEY-----"
	repo := &mockSSHRepo{
		getDefaultFn: func(id uuid.UUID) (*types.SSHKey, error) {
			assert.Equal(t, org, id)
			return &types.SSHKey{
				Host:                sp(host),
				User:                sp(user),
				Port:                ip(2222),
				PrivateKeyEncrypted: sp(priv),
			}, nil
		},
	}
	svc := service.NewSSHKeyServiceForTest(nil, context.Background(), logger.NewLogger(), repo)

	cfg, err := svc.GetSSHConfigForOrganization(org)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, host, cfg.Host)
	assert.Equal(t, user, cfg.User)
	assert.Equal(t, uint(2222), cfg.Port)
	assert.Equal(t, priv, cfg.PrivateKey)
	assert.Empty(t, repo.promoteCalls)
}

func TestGetSSHConfigForOrganization_promotesActiveWhenNoDefault(t *testing.T) {
	org := uuid.New()
	keyID := uuid.New()
	host := "10.0.0.2"
	repo := &mockSSHRepo{
		getDefaultFn: func(uuid.UUID) (*types.SSHKey, error) { return nil, sql.ErrNoRows },
		getActiveFn: func(uuid.UUID) (*types.SSHKey, error) {
			return &types.SSHKey{
				ID:                  keyID,
				Host:                sp(host),
				User:                sp("root"),
				PasswordEncrypted:   sp("secret"),
				PrivateKeyEncrypted: nil,
			}, nil
		},
	}
	svc := service.NewSSHKeyServiceForTest(nil, context.Background(), logger.NewLogger(), repo)

	cfg, err := svc.GetSSHConfigForOrganization(org)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, []uuid.UUID{keyID}, repo.promoteCalls)
	assert.Equal(t, host, cfg.Host)
	assert.Equal(t, "secret", cfg.Password)
	assert.Empty(t, cfg.PrivateKey)
}

func TestGetSSHConfigForOrganization_getDefaultFails(t *testing.T) {
	org := uuid.New()
	dbErr := errors.New("db exploded")
	repo := &mockSSHRepo{
		getDefaultFn: func(uuid.UUID) (*types.SSHKey, error) { return nil, dbErr },
	}
	svc := service.NewSSHKeyServiceForTest(nil, context.Background(), logger.NewLogger(), repo)

	cfg, err := svc.GetSSHConfigForOrganization(org)
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get SSH key for organization")
}

func TestGetSSHConfigForOrganization_noActiveKeys(t *testing.T) {
	org := uuid.New()
	repo := &mockSSHRepo{
		getDefaultFn: func(uuid.UUID) (*types.SSHKey, error) { return nil, sql.ErrNoRows },
		getActiveFn:  func(uuid.UUID) (*types.SSHKey, error) { return nil, sql.ErrNoRows },
	}
	svc := service.NewSSHKeyServiceForTest(nil, context.Background(), logger.NewLogger(), repo)

	cfg, err := svc.GetSSHConfigForOrganization(org)
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no server configured for organization")
}

func TestGetSSHConfigForOrganization_getActiveFails(t *testing.T) {
	org := uuid.New()
	dbErr := errors.New("lookup failed")
	repo := &mockSSHRepo{
		getDefaultFn: func(uuid.UUID) (*types.SSHKey, error) { return nil, sql.ErrNoRows },
		getActiveFn:  func(uuid.UUID) (*types.SSHKey, error) { return nil, dbErr },
	}
	svc := service.NewSSHKeyServiceForTest(nil, context.Background(), logger.NewLogger(), repo)

	cfg, err := svc.GetSSHConfigForOrganization(org)
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get SSH key for organization")
}

func TestGetSSHConfigForKey_nil(t *testing.T) {
	svc := service.NewSSHKeyServiceForTest(nil, context.Background(), logger.NewLogger(), &mockSSHRepo{})

	cfg, err := svc.GetSSHConfigForKey(nil)
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ssh key is nil")
}

func TestGetSSHConfigForKey_mapsPointersAndNegativePortDefault(t *testing.T) {
	svc := service.NewSSHKeyServiceForTest(nil, context.Background(), logger.NewLogger(), &mockSSHRepo{})
	key := &types.SSHKey{
		Host:                sp("1.2.3.4"),
		ProxyHost:           sp("jump.example"),
		User:                sp("deploy"),
		Port:                ip(-9),
		PrivateKeyEncrypted: sp("-----BEGIN RSA PRIVATE KEY-----\nX\n-----END RSA PRIVATE KEY-----"),
		PasswordEncrypted:   sp("pw"),
	}

	cfg, err := svc.GetSSHConfigForKey(key)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "1.2.3.4", cfg.Host)
	assert.Equal(t, "jump.example", cfg.ProxyHost)
	assert.Equal(t, "deploy", cfg.User)
	assert.Equal(t, uint(22), cfg.Port)
	assert.NotEmpty(t, cfg.PrivateKey)
	assert.Equal(t, "pw", cfg.Password)
}

func TestGetSSHConfigForKey_nilPointersUseDefaults(t *testing.T) {
	svc := service.NewSSHKeyServiceForTest(nil, context.Background(), logger.NewLogger(), &mockSSHRepo{})
	key := &types.SSHKey{}

	cfg, err := svc.GetSSHConfigForKey(key)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Empty(t, cfg.Host)
	assert.Equal(t, uint(22), cfg.Port)
}

func TestNewSSHKeyService_usesDefaultRepository(t *testing.T) {
	setup := testutils.NewTestSetup()
	svc := service.NewSSHKeyService(setup.Store, setup.Ctx, logger.NewLogger())
	require.NotNil(t, svc)
}
