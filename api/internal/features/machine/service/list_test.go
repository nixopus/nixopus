package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/service"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	sshpkg "github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cryptossh "golang.org/x/crypto/ssh"
)

// mockListRepo implements storage.ListRepository.
type mockListRepo struct {
	listFn          func(orgID uuid.UUID, params types.MachineListParams) ([]types.MachineResponse, int, error)
	setDefaultFn    func(orgID uuid.UUID, machineID uuid.UUID) (*uuid.UUID, error)
	getByIDAndOrgFn func(machineID uuid.UUID, orgID uuid.UUID) (*shared_types.SSHKey, error)
}

type mockSSHChecker struct {
	getConfigFn func() (*sshpkg.SSH, error)
	newSessFn   func(id string) (*cryptossh.Session, error)
}

func (m *mockSSHChecker) GetSSHConfig() (*sshpkg.SSH, error) {
	if m.getConfigFn != nil {
		return m.getConfigFn()
	}
	return &sshpkg.SSH{Host: "127.0.0.1"}, nil
}

func (m *mockSSHChecker) NewSessionWithRetry(id string) (*cryptossh.Session, error) {
	if m.newSessFn != nil {
		return m.newSessFn(id)
	}
	return nil, nil
}

func (m *mockListRepo) ListMachinesByOrganizationID(orgID uuid.UUID, params types.MachineListParams) ([]types.MachineResponse, int, error) {
	if m.listFn != nil {
		return m.listFn(orgID, params)
	}
	return []types.MachineResponse{}, 0, nil
}

func (m *mockListRepo) SetDefaultMachine(orgID uuid.UUID, machineID uuid.UUID) (*uuid.UUID, error) {
	if m.setDefaultFn != nil {
		return m.setDefaultFn(orgID, machineID)
	}
	return nil, nil
}

func (m *mockListRepo) GetMachineByIDAndOrgID(machineID uuid.UUID, orgID uuid.UUID) (*shared_types.SSHKey, error) {
	if m.getByIDAndOrgFn != nil {
		return m.getByIDAndOrgFn(machineID, orgID)
	}
	return &shared_types.SSHKey{ID: machineID}, nil
}

func TestNewListService(t *testing.T) {
	svc := service.NewListService(&mockListRepo{}, logger.NewLogger(), context.Background())
	assert.NotNil(t, svc)
}

func TestListService_ListMachines_DefaultParams(t *testing.T) {
	var capturedParams types.MachineListParams
	repo := &mockListRepo{
		listFn: func(_ uuid.UUID, p types.MachineListParams) ([]types.MachineResponse, int, error) {
			capturedParams = p
			return []types.MachineResponse{}, 5, nil
		},
	}
	svc := service.NewListService(repo, logger.NewLogger(), context.Background())

	resp, err := svc.ListMachines(uuid.New(), types.MachineListParams{})
	require.NoError(t, err)
	assert.Equal(t, 1, capturedParams.Page)
	assert.Equal(t, 10, capturedParams.PageSize)
	assert.Equal(t, "created_at", capturedParams.SortBy)
	assert.Equal(t, "desc", capturedParams.SortOrder)
	assert.Equal(t, 5, resp.Data.TotalCount)
}

func TestListService_ListMachines_ExplicitParams(t *testing.T) {
	var capturedParams types.MachineListParams
	repo := &mockListRepo{
		listFn: func(_ uuid.UUID, p types.MachineListParams) ([]types.MachineResponse, int, error) {
			capturedParams = p
			return make([]types.MachineResponse, 3), 3, nil
		},
	}
	svc := service.NewListService(repo, logger.NewLogger(), context.Background())

	params := types.MachineListParams{Page: 2, PageSize: 5, SortBy: "name", SortOrder: "asc"}
	resp, err := svc.ListMachines(uuid.New(), params)
	require.NoError(t, err)
	assert.Equal(t, 2, capturedParams.Page)
	assert.Equal(t, 5, capturedParams.PageSize)
	assert.Equal(t, "name", capturedParams.SortBy)
	assert.Equal(t, "asc", capturedParams.SortOrder)
	assert.Len(t, resp.Data.Servers, 3)
}

func TestListService_ListMachines_NilBecomesEmpty(t *testing.T) {
	repo := &mockListRepo{
		listFn: func(_ uuid.UUID, _ types.MachineListParams) ([]types.MachineResponse, int, error) {
			return nil, 0, nil
		},
	}
	svc := service.NewListService(repo, logger.NewLogger(), context.Background())

	resp, err := svc.ListMachines(uuid.New(), types.MachineListParams{})
	require.NoError(t, err)
	assert.NotNil(t, resp.Data.Servers)
	assert.Len(t, resp.Data.Servers, 0)
}

func TestListService_ListMachines_StorageError(t *testing.T) {
	repo := &mockListRepo{
		listFn: func(_ uuid.UUID, _ types.MachineListParams) ([]types.MachineResponse, int, error) {
			return nil, 0, fmt.Errorf("db error")
		},
	}
	svc := service.NewListService(repo, logger.NewLogger(), context.Background())

	_, err := svc.ListMachines(uuid.New(), types.MachineListParams{})
	require.Error(t, err)
}

func TestListService_SetDefaultMachine_StorageError(t *testing.T) {
	repo := &mockListRepo{
		setDefaultFn: func(_ uuid.UUID, _ uuid.UUID) (*uuid.UUID, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svc := service.NewListService(repo, logger.NewLogger(), context.Background())

	_, err := svc.SetDefaultMachine(uuid.New(), uuid.New())
	require.Error(t, err)
}

func TestListService_SetDefaultMachine_GetByIDError(t *testing.T) {
	repo := &mockListRepo{
		setDefaultFn: func(_ uuid.UUID, _ uuid.UUID) (*uuid.UUID, error) {
			return nil, nil // no old default
		},
		getByIDAndOrgFn: func(_ uuid.UUID, _ uuid.UUID) (*shared_types.SSHKey, error) {
			return nil, fmt.Errorf("key not found")
		},
	}
	svc := service.NewListService(repo, logger.NewLogger(), context.Background())

	_, err := svc.SetDefaultMachine(uuid.New(), uuid.New())
	require.Error(t, err)
}

func TestListService_SetDefaultMachine_SuccessNoOldDefault(t *testing.T) {
	machineID := uuid.New()
	repo := &mockListRepo{
		setDefaultFn: func(_ uuid.UUID, _ uuid.UUID) (*uuid.UUID, error) {
			return nil, nil
		},
		getByIDAndOrgFn: func(id uuid.UUID, _ uuid.UUID) (*shared_types.SSHKey, error) {
			return &shared_types.SSHKey{ID: id}, nil
		},
	}
	svc := service.NewListService(repo, logger.NewLogger(), context.Background())

	key, err := svc.SetDefaultMachine(uuid.New(), machineID)
	require.NoError(t, err)
	assert.Equal(t, machineID, key.ID)
}

func TestListService_SetDefaultMachine_SuccessWithOldDefault(t *testing.T) {
	machineID := uuid.New()
	oldDefault := uuid.New()
	repo := &mockListRepo{
		setDefaultFn: func(_ uuid.UUID, _ uuid.UUID) (*uuid.UUID, error) {
			return &oldDefault, nil
		},
		getByIDAndOrgFn: func(id uuid.UUID, _ uuid.UUID) (*shared_types.SSHKey, error) {
			return &shared_types.SSHKey{ID: id}, nil
		},
	}
	svc := service.NewListService(repo, logger.NewLogger(), context.Background())

	key, err := svc.SetDefaultMachine(uuid.New(), machineID)
	require.NoError(t, err)
	assert.Equal(t, machineID, key.ID)
}

func TestListService_CheckSSHConnection_Error(t *testing.T) {
	// With a context without any configured org SSH key, GetSSHManagerForOrganization
	// will fail, returning the error response (not an error return, but status="error").
	svc := service.NewListService(&mockListRepo{}, logger.NewLogger(), context.Background())
	resp, err := svc.CheckSSHConnection(uuid.New())
	require.NoError(t, err) // CheckSSHConnection never returns an error
	assert.Equal(t, "error", resp.Status)
	assert.False(t, resp.Connected)
	assert.False(t, resp.IsConfigured)
}

func TestListService_CheckSSHConnection_NotConfigured_ErrorFromGetConfig(t *testing.T) {
	svc := service.NewListService(&mockListRepo{}, logger.NewLogger(), context.Background())
	svc.SetSSHManagerProviderForTest(func(_ context.Context, _ uuid.UUID) (service.SSHConnectionChecker, error) {
		return &mockSSHChecker{
			getConfigFn: func() (*sshpkg.SSH, error) { return nil, fmt.Errorf("no config") },
		}, nil
	})
	resp, err := svc.CheckSSHConnection(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "not_configured", resp.Status)
	assert.False(t, resp.Connected)
	assert.False(t, resp.IsConfigured)
}

func TestListService_CheckSSHConnection_NotConfigured_EmptyHost(t *testing.T) {
	svc := service.NewListService(&mockListRepo{}, logger.NewLogger(), context.Background())
	svc.SetSSHManagerProviderForTest(func(_ context.Context, _ uuid.UUID) (service.SSHConnectionChecker, error) {
		return &mockSSHChecker{
			getConfigFn: func() (*sshpkg.SSH, error) { return &sshpkg.SSH{Host: ""}, nil },
		}, nil
	})
	resp, err := svc.CheckSSHConnection(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "not_configured", resp.Status)
}

func TestListService_CheckSSHConnection_NotConfigured_NilConfig(t *testing.T) {
	svc := service.NewListService(&mockListRepo{}, logger.NewLogger(), context.Background())
	svc.SetSSHManagerProviderForTest(func(_ context.Context, _ uuid.UUID) (service.SSHConnectionChecker, error) {
		return &mockSSHChecker{
			getConfigFn: func() (*sshpkg.SSH, error) { return nil, nil },
		}, nil
	})
	resp, err := svc.CheckSSHConnection(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "not_configured", resp.Status)
}

func TestListService_CheckSSHConnection_Disconnected(t *testing.T) {
	svc := service.NewListService(&mockListRepo{}, logger.NewLogger(), context.Background())
	svc.SetSSHManagerProviderForTest(func(_ context.Context, _ uuid.UUID) (service.SSHConnectionChecker, error) {
		return &mockSSHChecker{
			getConfigFn: func() (*sshpkg.SSH, error) { return &sshpkg.SSH{Host: "10.0.0.1"}, nil },
			newSessFn:   func(_ string) (*cryptossh.Session, error) { return nil, fmt.Errorf("dial failed") },
		}, nil
	})
	resp, err := svc.CheckSSHConnection(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "disconnected", resp.Status)
	assert.False(t, resp.Connected)
	assert.True(t, resp.IsConfigured)
}

func TestListService_CheckSSHConnection_Connected(t *testing.T) {
	svc := service.NewListService(&mockListRepo{}, logger.NewLogger(), context.Background())
	svc.SetSSHManagerProviderForTest(func(_ context.Context, _ uuid.UUID) (service.SSHConnectionChecker, error) {
		return &mockSSHChecker{
			getConfigFn: func() (*sshpkg.SSH, error) { return &sshpkg.SSH{Host: "10.0.0.1"}, nil },
			newSessFn:   func(_ string) (*cryptossh.Session, error) { return nil, nil },
		}, nil
	})
	resp, err := svc.CheckSSHConnection(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "connected", resp.Status)
	assert.True(t, resp.Connected)
	assert.True(t, resp.IsConfigured)
}
