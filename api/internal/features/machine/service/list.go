package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	sshpkg "github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	cryptossh "golang.org/x/crypto/ssh"
)

// SSHConnectionChecker is the minimal SSH manager interface required by ListService.
type SSHConnectionChecker interface {
	GetSSHConfig() (*sshpkg.SSH, error)
	NewSessionWithRetry(id string) (*cryptossh.Session, error)
}

type ListService struct {
	storage          storage.ListRepository
	logger           logger.Logger
	ctx              context.Context
	sshMgrProviderFn func(ctx context.Context, orgID uuid.UUID) (SSHConnectionChecker, error) // nil -> production
}

func NewListService(storage storage.ListRepository, l logger.Logger, ctx context.Context) *ListService {
	return &ListService{storage: storage, logger: l, ctx: ctx}
}

// SetSSHManagerProviderForTest injects an SSH manager provider for tests.
func (s *ListService) SetSSHManagerProviderForTest(fn func(ctx context.Context, orgID uuid.UUID) (SSHConnectionChecker, error)) {
	s.sshMgrProviderFn = fn
}

func (s *ListService) ListMachines(orgID uuid.UUID, params types.MachineListParams) (*types.ListMachinesResponse, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	machines, totalCount, err := s.storage.ListMachinesByOrganizationID(orgID, params)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("machine service: ListMachines: %v", err), fmt.Sprintf("org_id=%s", orgID))
		return nil, err
	}

	if machines == nil {
		machines = []types.MachineResponse{}
	}

	return &types.ListMachinesResponse{
		Status:  "success",
		Message: "Servers fetched successfully",
		Data: types.ListMachinesResponseData{
			Servers:    machines,
			TotalCount: totalCount,
			Page:       params.Page,
			PageSize:   params.PageSize,
			SortBy:     params.SortBy,
			SortOrder:  params.SortOrder,
			Search:     params.Search,
			Status:     params.Status,
			IsActive:   params.IsActive,
		},
	}, nil
}

func (s *ListService) SetDefaultMachine(orgID uuid.UUID, machineID uuid.UUID) (*shared_types.SSHKey, error) {
	oldDefaultID, err := s.storage.SetDefaultMachine(orgID, machineID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("machine service: SetDefaultMachine: %v", err), fmt.Sprintf("org_id=%s machine_id=%s", orgID, machineID))
		return nil, err
	}

	if oldDefaultID != nil {
		sshpkg.InvalidateServerManagerCache(*oldDefaultID)
	}

	key, err := s.storage.GetMachineByIDAndOrgID(machineID, orgID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("machine service: GetMachineByIDAndOrgID: %v", err), fmt.Sprintf("org_id=%s machine_id=%s", orgID, machineID))
		return nil, err
	}
	return key, nil
}

func (s *ListService) CheckSSHConnection(orgID uuid.UUID) (*types.SSHConnectionStatusResponse, error) {
	sshProvider := s.sshMgrProviderFn
	if sshProvider == nil {
		sshProvider = func(ctx context.Context, orgID uuid.UUID) (SSHConnectionChecker, error) {
			return sshpkg.GetSSHManagerForOrganization(ctx, orgID)
		}
	}
	sshMgr, err := sshProvider(s.ctx, orgID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("machine service: CheckSSHConnection ssh manager: %v", err), fmt.Sprintf("org_id=%s", orgID))
		return &types.SSHConnectionStatusResponse{
			Status:       "error",
			Connected:    false,
			Message:      "Failed to get SSH manager",
			IsConfigured: false,
		}, nil
	}

	sshConfig, err := sshMgr.GetSSHConfig()
	if err != nil || sshConfig == nil || sshConfig.Host == "" {
		return &types.SSHConnectionStatusResponse{
			Status:       "not_configured",
			Connected:    false,
			Message:      "SSH is not configured for this organization",
			IsConfigured: false,
		}, nil
	}

	session, err := sshMgr.NewSessionWithRetry("")
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("machine service: CheckSSHConnection session: %v", err), fmt.Sprintf("org_id=%s", orgID))
		return &types.SSHConnectionStatusResponse{
			Status:       "disconnected",
			Connected:    false,
			Message:      "Unable to connect to SSH server",
			IsConfigured: true,
		}, nil
	}
	if session != nil {
		_ = session.Close()
	}

	return &types.SSHConnectionStatusResponse{
		Status:       "connected",
		Connected:    true,
		Message:      "SSH connection is active",
		IsConfigured: true,
	}, nil
}
