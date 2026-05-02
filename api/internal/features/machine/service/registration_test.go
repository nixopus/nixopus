package service_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/service"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	api_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	cryptossh "golang.org/x/crypto/ssh"
)

func TestNoOpBillingChecker_CanProvision(t *testing.T) {
	checker := &service.NoOpBillingChecker{}
	err := checker.CanProvision(uuid.New())
	assert.NoError(t, err)
}

func TestNewRegistrationService_NilBillingChecker(t *testing.T) {
	// nil billing checker should be replaced with NoOpBillingChecker internally
	svc := service.NewRegistrationService(nil, nil, nil, logger.NewLogger(), context.Background())
	assert.NotNil(t, svc)
}

func TestNewRegistrationService_WithChecker(t *testing.T) {
	checker := &service.NoOpBillingChecker{}
	svc := service.NewRegistrationService(nil, nil, checker, logger.NewLogger(), context.Background())
	assert.NotNil(t, svc)
}

func TestRegistrationService_RenameMachine_EmptyName(t *testing.T) {
	svc := service.NewRegistrationService(nil, nil, nil, logger.NewLogger(), context.Background())
	_, err := svc.RenameMachine(uuid.New(), uuid.New(), "")
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrNameRequired)
}

func TestRegistrationService_RenameMachine_WhitespaceName(t *testing.T) {
	svc := service.NewRegistrationService(nil, nil, nil, logger.NewLogger(), context.Background())
	_, err := svc.RenameMachine(uuid.New(), uuid.New(), "   ")
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrNameRequired)
}

func TestRegistrationService_RenameMachine_TooLongName(t *testing.T) {
	svc := service.NewRegistrationService(nil, nil, nil, logger.NewLogger(), context.Background())
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := svc.RenameMachine(uuid.New(), uuid.New(), string(longName))
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrNameTooLong)
}

func TestRegistrationService_RenameMachine_ExactlyMaxLength_PassesValidation(t *testing.T) {
	// 255 chars is the boundary — should pass the length check (len <= 255 is ok)
	// We verify no validation errors are returned by testing 255 == max, which
	// passes the > 255 guard and proceeds to the storage call. With nil storage it panics;
	// we catch that to confirm validation itself passed.
	svc := service.NewRegistrationService(nil, nil, nil, logger.NewLogger(), context.Background())
	name255 := make([]byte, 255)
	for i := range name255 {
		name255[i] = 'a'
	}
	assert.Panics(t, func() {
		//nolint:errcheck
		svc.RenameMachine(uuid.New(), uuid.New(), string(name255))
	})
}

// ---------- mockRegistrationRepo ----------

type mockRegistrationRepo struct {
	hostPortExistsFn           func(orgID uuid.UUID, host string, port int) (bool, error)
	runInTxFn                  func(fn func(bun.Tx) error) error
	insertSSHKeyTxFn           func(tx bun.Tx, key *api_types.SSHKey) error
	insertProvisionDetailsTxFn func(tx bun.Tx, userID, orgID, sshKeyID uuid.UUID, provType string, step api_types.ProvisionStep) error
	getSSHKeyByIDFn            func(id, orgID uuid.UUID) (*api_types.SSHKey, error)
	getSSHKeyStatusFn          func(id, orgID uuid.UUID) (bool, *time.Time, error)
	hasActiveAppServersFn      func(sshKeyID uuid.UUID) (bool, error)
	softDeleteSSHKeyFn         func(sshKeyID uuid.UUID) error
	updateMachineNameFn        func(sshKeyID uuid.UUID, name string) error
	markMachineActiveFn        func(sshKeyID uuid.UUID) error
	markMachineInactiveFn      func(sshKeyID uuid.UUID) error
}

func (m *mockRegistrationRepo) HostPortExists(orgID uuid.UUID, host string, port int) (bool, error) {
	if m.hostPortExistsFn != nil {
		return m.hostPortExistsFn(orgID, host, port)
	}
	return false, nil
}
func (m *mockRegistrationRepo) RunInTx(fn func(bun.Tx) error) error {
	if m.runInTxFn != nil {
		return m.runInTxFn(fn)
	}
	return fn(bun.Tx{})
}
func (m *mockRegistrationRepo) InsertSSHKeyTx(tx bun.Tx, key *api_types.SSHKey) error {
	if m.insertSSHKeyTxFn != nil {
		return m.insertSSHKeyTxFn(tx, key)
	}
	return nil
}
func (m *mockRegistrationRepo) InsertProvisionDetailsTx(tx bun.Tx, userID, orgID, sshKeyID uuid.UUID, provType string, step api_types.ProvisionStep) error {
	if m.insertProvisionDetailsTxFn != nil {
		return m.insertProvisionDetailsTxFn(tx, userID, orgID, sshKeyID, provType, step)
	}
	return nil
}
func (m *mockRegistrationRepo) GetSSHKeyByID(id, orgID uuid.UUID) (*api_types.SSHKey, error) {
	if m.getSSHKeyByIDFn != nil {
		return m.getSSHKeyByIDFn(id, orgID)
	}
	return &api_types.SSHKey{ID: id}, nil
}
func (m *mockRegistrationRepo) GetSSHKeyStatus(id, orgID uuid.UUID) (bool, *time.Time, error) {
	if m.getSSHKeyStatusFn != nil {
		return m.getSSHKeyStatusFn(id, orgID)
	}
	return true, nil, nil
}
func (m *mockRegistrationRepo) HasActiveAppServers(sshKeyID uuid.UUID) (bool, error) {
	if m.hasActiveAppServersFn != nil {
		return m.hasActiveAppServersFn(sshKeyID)
	}
	return false, nil
}
func (m *mockRegistrationRepo) SoftDeleteSSHKey(sshKeyID uuid.UUID) error {
	if m.softDeleteSSHKeyFn != nil {
		return m.softDeleteSSHKeyFn(sshKeyID)
	}
	return nil
}
func (m *mockRegistrationRepo) UpdateMachineName(sshKeyID uuid.UUID, name string) error {
	if m.updateMachineNameFn != nil {
		return m.updateMachineNameFn(sshKeyID, name)
	}
	return nil
}
func (m *mockRegistrationRepo) MarkMachineActive(sshKeyID uuid.UUID) error {
	if m.markMachineActiveFn != nil {
		return m.markMachineActiveFn(sshKeyID)
	}
	return nil
}
func (m *mockRegistrationRepo) MarkMachineInactive(sshKeyID uuid.UUID) error {
	if m.markMachineInactiveFn != nil {
		return m.markMachineInactiveFn(sshKeyID)
	}
	return nil
}

func newTestRegistrationService(repo service.RegistrationRepository) *service.RegistrationService {
	return service.NewRegistrationServiceWith(repo, nil, nil, logger.NewLogger(), context.Background())
}

// ---------- CreateMachine ----------

func TestRegistrationService_CreateMachine_HostPortExistsError(t *testing.T) {
	repo := &mockRegistrationRepo{
		hostPortExistsFn: func(_ uuid.UUID, _ string, _ int) (bool, error) {
			return false, fmt.Errorf("db error")
		},
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.CreateMachine(uuid.New(), uuid.New(), types.CreateMachineRequest{Host: "1.2.3.4", Port: 22})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check host uniqueness")
}

func TestRegistrationService_CreateMachine_DuplicateHost(t *testing.T) {
	repo := &mockRegistrationRepo{
		hostPortExistsFn: func(_ uuid.UUID, _ string, _ int) (bool, error) { return true, nil },
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.CreateMachine(uuid.New(), uuid.New(), types.CreateMachineRequest{Host: "1.2.3.4"})
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrDuplicateHost)
}

func TestRegistrationService_CreateMachine_InsertKeyTxError(t *testing.T) {
	repo := &mockRegistrationRepo{
		hostPortExistsFn: func(_ uuid.UUID, _ string, _ int) (bool, error) { return false, nil },
		runInTxFn: func(fn func(bun.Tx) error) error {
			return fn(bun.Tx{}) // propagate the fn error
		},
		insertSSHKeyTxFn: func(_ bun.Tx, _ *api_types.SSHKey) error {
			return fmt.Errorf("insert key failed")
		},
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.CreateMachine(uuid.New(), uuid.New(), types.CreateMachineRequest{Host: "1.2.3.4", Name: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert ssh key")
}

func TestRegistrationService_CreateMachine_InsertProvisionTxError(t *testing.T) {
	repo := &mockRegistrationRepo{
		hostPortExistsFn: func(_ uuid.UUID, _ string, _ int) (bool, error) { return false, nil },
		runInTxFn: func(fn func(bun.Tx) error) error {
			return fn(bun.Tx{})
		},
		insertSSHKeyTxFn: func(_ bun.Tx, _ *api_types.SSHKey) error { return nil },
		insertProvisionDetailsTxFn: func(_ bun.Tx, _, _, _ uuid.UUID, _ string, _ api_types.ProvisionStep) error {
			return fmt.Errorf("provision insert failed")
		},
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.CreateMachine(uuid.New(), uuid.New(), types.CreateMachineRequest{Host: "1.2.3.4", Name: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert provision details")
}

func TestRegistrationService_CreateMachine_Success(t *testing.T) {
	repo := &mockRegistrationRepo{
		hostPortExistsFn: func(_ uuid.UUID, _ string, _ int) (bool, error) { return false, nil },
	}
	svc := newTestRegistrationService(repo)
	resp, err := svc.CreateMachine(uuid.New(), uuid.New(), types.CreateMachineRequest{
		Host: "1.2.3.4",
		Name: "my-server",
	})
	require.NoError(t, err)
	assert.Equal(t, "my-server", resp.Name)
	assert.Equal(t, "1.2.3.4", resp.Host)
	assert.Equal(t, 22, resp.Port)     // default port
	assert.Equal(t, "root", resp.User) // default user
}

func TestRegistrationService_CreateMachine_WithExplicitPortAndUser(t *testing.T) {
	repo := &mockRegistrationRepo{
		hostPortExistsFn: func(_ uuid.UUID, _ string, _ int) (bool, error) { return false, nil },
	}
	svc := newTestRegistrationService(repo)
	resp, err := svc.CreateMachine(uuid.New(), uuid.New(), types.CreateMachineRequest{
		Host: "1.2.3.4",
		Name: "my-server",
		Port: 2222,
		User: "ubuntu",
	})
	require.NoError(t, err)
	assert.Equal(t, 2222, resp.Port)
	assert.Equal(t, "ubuntu", resp.User)
}

// ---------- DeleteMachine ----------

func TestRegistrationService_DeleteMachine_GetKeyError(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_ uuid.UUID, _ uuid.UUID) (*api_types.SSHKey, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svc := newTestRegistrationService(repo)
	err := svc.DeleteMachine(uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "machine not found")
}

func TestRegistrationService_DeleteMachine_HasAppsError(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_ uuid.UUID, _ uuid.UUID) (*api_types.SSHKey, error) {
			return &api_types.SSHKey{}, nil
		},
		hasActiveAppServersFn: func(_ uuid.UUID) (bool, error) {
			return false, fmt.Errorf("db error")
		},
	}
	svc := newTestRegistrationService(repo)
	err := svc.DeleteMachine(uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check app servers")
}

func TestRegistrationService_DeleteMachine_HasApps(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn:       func(_ uuid.UUID, _ uuid.UUID) (*api_types.SSHKey, error) { return &api_types.SSHKey{}, nil },
		hasActiveAppServersFn: func(_ uuid.UUID) (bool, error) { return true, nil },
	}
	svc := newTestRegistrationService(repo)
	err := svc.DeleteMachine(uuid.New(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineHasApps)
}

func TestRegistrationService_DeleteMachine_Success(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn:       func(_ uuid.UUID, _ uuid.UUID) (*api_types.SSHKey, error) { return &api_types.SSHKey{}, nil },
		hasActiveAppServersFn: func(_ uuid.UUID) (bool, error) { return false, nil },
		softDeleteSSHKeyFn:    func(_ uuid.UUID) error { return nil },
	}
	svc := newTestRegistrationService(repo)
	err := svc.DeleteMachine(uuid.New(), uuid.New())
	require.NoError(t, err)
}

// ---------- GetSSHStatus ----------

func TestRegistrationService_GetSSHStatus_Error(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyStatusFn: func(_ uuid.UUID, _ uuid.UUID) (bool, *time.Time, error) {
			return false, nil, fmt.Errorf("not found")
		},
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.GetSSHStatus(uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "machine not found")
}

func TestRegistrationService_GetSSHStatus_Active_NoLastUsed(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyStatusFn: func(_ uuid.UUID, _ uuid.UUID) (bool, *time.Time, error) {
			return true, nil, nil
		},
	}
	svc := newTestRegistrationService(repo)
	resp, err := svc.GetSSHStatus(uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.True(t, resp.IsActive)
	assert.Empty(t, resp.LastUsedAt)
}

func TestRegistrationService_GetSSHStatus_Active_WithLastUsed(t *testing.T) {
	now := time.Now()
	repo := &mockRegistrationRepo{
		getSSHKeyStatusFn: func(_ uuid.UUID, _ uuid.UUID) (bool, *time.Time, error) {
			return true, &now, nil
		},
	}
	svc := newTestRegistrationService(repo)
	resp, err := svc.GetSSHStatus(uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.NotEmpty(t, resp.LastUsedAt)
}

// ---------- RenameMachine (storage paths) ----------

func TestRegistrationService_RenameMachine_GetKeyNotFound(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_ uuid.UUID, _ uuid.UUID) (*api_types.SSHKey, error) {
			return nil, sql.ErrNoRows
		},
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.RenameMachine(uuid.New(), uuid.New(), "valid-name")
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineNotFound)
}

func TestRegistrationService_RenameMachine_GetKeyOtherError(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_ uuid.UUID, _ uuid.UUID) (*api_types.SSHKey, error) {
			return nil, fmt.Errorf("connection reset")
		},
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.RenameMachine(uuid.New(), uuid.New(), "valid-name")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get machine")
}

func TestRegistrationService_RenameMachine_UpdateNotFound(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_ uuid.UUID, _ uuid.UUID) (*api_types.SSHKey, error) {
			return &api_types.SSHKey{}, nil
		},
		updateMachineNameFn: func(_ uuid.UUID, _ string) error { return sql.ErrNoRows },
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.RenameMachine(uuid.New(), uuid.New(), "valid-name")
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineNotFound)
}

func TestRegistrationService_RenameMachine_UpdateOtherError(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn:     func(_ uuid.UUID, _ uuid.UUID) (*api_types.SSHKey, error) { return &api_types.SSHKey{}, nil },
		updateMachineNameFn: func(_ uuid.UUID, _ string) error { return fmt.Errorf("db error") },
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.RenameMachine(uuid.New(), uuid.New(), "valid-name")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to rename machine")
}

func TestRegistrationService_RenameMachine_Success(t *testing.T) {
	machineID := uuid.New()
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn:     func(_ uuid.UUID, _ uuid.UUID) (*api_types.SSHKey, error) { return &api_types.SSHKey{}, nil },
		updateMachineNameFn: func(_ uuid.UUID, _ string) error { return nil },
	}
	svc := newTestRegistrationService(repo)
	resp, err := svc.RenameMachine(uuid.New(), machineID, "  my server  ")
	require.NoError(t, err)
	assert.Equal(t, "my server", resp.Name) // trimmed
	assert.Equal(t, machineID.String(), resp.ID)
}

type mockRegSSHSession struct {
	runFn   func(cmd string) error
	closeFn func() error
}

func (m *mockRegSSHSession) Run(cmd string) error {
	if m.runFn != nil {
		return m.runFn(cmd)
	}
	return nil
}

func (m *mockRegSSHSession) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

type mockRegSSHClient struct {
	newSessionFn func() (service.RegistrationSSHSession, error)
	closeFn      func() error
}

func (m *mockRegSSHClient) NewSession() (service.RegistrationSSHSession, error) {
	if m.newSessionFn != nil {
		return m.newSessionFn()
	}
	return &mockRegSSHSession{}, nil
}

func (m *mockRegSSHClient) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func testSigner(t *testing.T) cryptossh.Signer {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	signer, err := cryptossh.NewSignerFromKey(key)
	require.NoError(t, err)
	return signer
}

func TestRegistrationService_VerifyMachine_NotFound(t *testing.T) {
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_, _ uuid.UUID) (*api_types.SSHKey, error) { return nil, sql.ErrNoRows },
	}
	svc := newTestRegistrationService(repo)
	_, err := svc.VerifyMachine(uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "machine not found")
}

func TestRegistrationService_VerifyMachine_MissingKeyData(t *testing.T) {
	host := "127.0.0.1"
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_, _ uuid.UUID) (*api_types.SSHKey, error) {
			return &api_types.SSHKey{Host: &host, PrivateKeyEncrypted: nil}, nil
		},
	}
	svc := newTestRegistrationService(repo)
	resp, err := svc.VerifyMachine(uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "failed", resp.Status)
	assert.False(t, resp.IsActive)
}

func TestRegistrationService_VerifyMachine_ParseKeyError(t *testing.T) {
	host := "127.0.0.1"
	priv := "invalid-private-key"
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_, _ uuid.UUID) (*api_types.SSHKey, error) {
			return &api_types.SSHKey{Host: &host, PrivateKeyEncrypted: &priv}, nil
		},
	}
	svc := newTestRegistrationService(repo)
	resp, err := svc.VerifyMachine(uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "failed", resp.Status)
}

func TestRegistrationService_VerifyMachine_DialError_MarksInactive(t *testing.T) {
	host := "127.0.0.1"
	priv := "unused"
	machineID := uuid.New()
	markedInactive := false
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(id, _ uuid.UUID) (*api_types.SSHKey, error) {
			assert.Equal(t, machineID, id)
			return &api_types.SSHKey{Host: &host, PrivateKeyEncrypted: &priv}, nil
		},
		markMachineInactiveFn: func(id uuid.UUID) error {
			assert.Equal(t, machineID, id)
			markedInactive = true
			return nil
		},
	}
	svc := newTestRegistrationService(repo)
	svc.SetParsePrivateKeyFnForTest(func(_ []byte) (cryptossh.Signer, error) { return testSigner(t), nil })
	svc.SetDialSSHFnForTest(func(_, _ string, _ *cryptossh.ClientConfig) (service.RegistrationSSHClient, error) {
		return nil, fmt.Errorf("dial failed")
	})
	resp, err := svc.VerifyMachine(uuid.New(), machineID)
	require.NoError(t, err)
	assert.Equal(t, "failed", resp.Status)
	assert.True(t, markedInactive)
}

func TestRegistrationService_VerifyMachine_NewSessionError_MarksInactive(t *testing.T) {
	host := "127.0.0.1"
	priv := "unused"
	machineID := uuid.New()
	markedInactive := false
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_, _ uuid.UUID) (*api_types.SSHKey, error) {
			return &api_types.SSHKey{Host: &host, PrivateKeyEncrypted: &priv}, nil
		},
		markMachineInactiveFn: func(_ uuid.UUID) error {
			markedInactive = true
			return nil
		},
	}
	svc := newTestRegistrationService(repo)
	svc.SetParsePrivateKeyFnForTest(func(_ []byte) (cryptossh.Signer, error) { return testSigner(t), nil })
	svc.SetDialSSHFnForTest(func(_, _ string, _ *cryptossh.ClientConfig) (service.RegistrationSSHClient, error) {
		return &mockRegSSHClient{
			newSessionFn: func() (service.RegistrationSSHSession, error) {
				return nil, fmt.Errorf("session failed")
			},
		}, nil
	})
	resp, err := svc.VerifyMachine(uuid.New(), machineID)
	require.NoError(t, err)
	assert.Equal(t, "failed", resp.Status)
	assert.True(t, markedInactive)
}

func TestRegistrationService_VerifyMachine_RunError_MarksInactive(t *testing.T) {
	host := "127.0.0.1"
	priv := "unused"
	machineID := uuid.New()
	markedInactive := false
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_, _ uuid.UUID) (*api_types.SSHKey, error) {
			return &api_types.SSHKey{Host: &host, PrivateKeyEncrypted: &priv}, nil
		},
		markMachineInactiveFn: func(_ uuid.UUID) error {
			markedInactive = true
			return nil
		},
	}
	svc := newTestRegistrationService(repo)
	svc.SetParsePrivateKeyFnForTest(func(_ []byte) (cryptossh.Signer, error) { return testSigner(t), nil })
	svc.SetDialSSHFnForTest(func(_, _ string, _ *cryptossh.ClientConfig) (service.RegistrationSSHClient, error) {
		return &mockRegSSHClient{
			newSessionFn: func() (service.RegistrationSSHSession, error) {
				return &mockRegSSHSession{
					runFn: func(_ string) error { return fmt.Errorf("run failed") },
				}, nil
			},
		}, nil
	})
	resp, err := svc.VerifyMachine(uuid.New(), machineID)
	require.NoError(t, err)
	assert.Equal(t, "failed", resp.Status)
	assert.True(t, markedInactive)
}

func TestRegistrationService_VerifyMachine_MarkActiveError(t *testing.T) {
	host := "127.0.0.1"
	priv := "unused"
	machineID := uuid.New()
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_, _ uuid.UUID) (*api_types.SSHKey, error) {
			return &api_types.SSHKey{Host: &host, PrivateKeyEncrypted: &priv}, nil
		},
		markMachineActiveFn: func(_ uuid.UUID) error { return fmt.Errorf("update failed") },
	}
	svc := newTestRegistrationService(repo)
	svc.SetParsePrivateKeyFnForTest(func(_ []byte) (cryptossh.Signer, error) { return testSigner(t), nil })
	svc.SetDialSSHFnForTest(func(_, _ string, _ *cryptossh.ClientConfig) (service.RegistrationSSHClient, error) {
		return &mockRegSSHClient{
			newSessionFn: func() (service.RegistrationSSHSession, error) {
				return &mockRegSSHSession{}, nil
			},
		}, nil
	})
	_, err := svc.VerifyMachine(uuid.New(), machineID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update machine status")
}

func TestRegistrationService_VerifyMachine_Success(t *testing.T) {
	host := "127.0.0.1"
	priv := "unused"
	machineID := uuid.New()
	repo := &mockRegistrationRepo{
		getSSHKeyByIDFn: func(_, _ uuid.UUID) (*api_types.SSHKey, error) {
			return &api_types.SSHKey{Host: &host, PrivateKeyEncrypted: &priv}, nil
		},
		markMachineActiveFn: func(_ uuid.UUID) error { return nil },
	}
	svc := newTestRegistrationService(repo)
	svc.SetParsePrivateKeyFnForTest(func(_ []byte) (cryptossh.Signer, error) { return testSigner(t), nil })
	svc.SetDialSSHFnForTest(func(_, _ string, _ *cryptossh.ClientConfig) (service.RegistrationSSHClient, error) {
		return &mockRegSSHClient{
			newSessionFn: func() (service.RegistrationSSHSession, error) {
				return &mockRegSSHSession{}, nil
			},
		}, nil
	})
	resp, err := svc.VerifyMachine(uuid.New(), machineID)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.True(t, resp.IsActive)
}
