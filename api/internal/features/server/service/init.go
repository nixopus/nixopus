package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/server/storage"
	"github.com/nixopus/nixopus/api/internal/features/server/types"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type ServerService struct {
	store   *shared_storage.Store
	ctx     context.Context
	logger  logger.Logger
	storage storage.ServerRepository
}

func NewServerService(store *shared_storage.Store, ctx context.Context, l logger.Logger) *ServerService {
	serverStorage := &storage.ServerStorage{DB: store.DB, Ctx: ctx}
	return &ServerService{
		store:   store,
		ctx:     ctx,
		logger:  l,
		storage: serverStorage,
	}
}

// ListServers retrieves a paginated, filtered, and sorted list of servers (SSH keys)
// for an organization with optional user_provision_details.
func (s *ServerService) ListServers(orgID uuid.UUID, params types.ServerListParams) (*types.ListServersResponse, error) {
	// Set defaults for pagination
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// Set default sorting
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// Call storage layer
	servers, totalCount, err := s.storage.ListServersByOrganizationID(orgID, params)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, err
	}

	// Return empty slice if nil
	if servers == nil {
		servers = []types.ServerResponse{}
	}

	return &types.ListServersResponse{
		Status:  "success",
		Message: "Servers fetched successfully",
		Data: types.ListServersResponseData{
			Servers:    servers,
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

// SetDefaultServer designates serverID as the org's active default server.
// Invalidates the old default's SSH manager cache entry after the DB transaction.
func (s *ServerService) SetDefaultServer(orgID uuid.UUID, serverID uuid.UUID) (*shared_types.SSHKey, error) {
	oldDefaultID, err := s.storage.SetDefaultServer(orgID, serverID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, err
	}

	if oldDefaultID != nil {
		ssh.InvalidateServerManagerCache(*oldDefaultID)
	}

	key, err := s.storage.GetServerByIDAndOrgID(serverID, orgID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), serverID.String())
		return nil, err
	}
	return key, nil
}

// CheckSSHConnection checks if SSH connection is available for the organization
func (s *ServerService) CheckSSHConnection(orgID uuid.UUID) (*types.SSHConnectionStatusResponse, error) {
	// Get SSH manager for the organization
	sshMgr, err := ssh.GetSSHManagerForOrganization(s.ctx, orgID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), orgID.String())
		return &types.SSHConnectionStatusResponse{
			Status:       "error",
			Connected:    false,
			Message:      "Failed to get SSH manager",
			IsConfigured: false,
		}, nil
	}

	// Check if SSH is configured
	sshConfig, err := sshMgr.GetSSHConfig()
	if err != nil || sshConfig == nil || sshConfig.Host == "" {
		return &types.SSHConnectionStatusResponse{
			Status:       "not_configured",
			Connected:    false,
			Message:      "SSH is not configured for this organization",
			IsConfigured: false,
		}, nil
	}

	// Test connection by creating (and immediately closing) a session on the
	// pooled connection.  Do NOT call client.Close() -- that would destroy the
	// shared pooled transport used by terminals & other features.
	session, err := sshMgr.NewSessionWithRetry("")
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), orgID.String())
		return &types.SSHConnectionStatusResponse{
			Status:       "disconnected",
			Connected:    false,
			Message:      "Unable to connect to SSH server",
			IsConfigured: true,
		}, nil
	}
	session.Close()

	return &types.SSHConnectionStatusResponse{
		Status:       "connected",
		Connected:    true,
		Message:      "SSH connection is active",
		IsConfigured: true,
	}, nil
}

// CheckSSHConnectionByServerID checks SSH connectivity for a specific server by ID.
func (s *ServerService) CheckSSHConnectionByServerID(orgID uuid.UUID, serverID uuid.UUID) (*types.SSHConnectionStatusResponse, error) {
	sshMgr, err := ssh.GetSSHManagerForServer(s.ctx, orgID, serverID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), serverID.String())
		return &types.SSHConnectionStatusResponse{
			Status:       "error",
			Connected:    false,
			Message:      "Failed to get SSH manager for server",
			IsConfigured: false,
		}, nil
	}

	sshConfig, err := sshMgr.GetSSHConfig()
	if err != nil || sshConfig == nil || sshConfig.Host == "" {
		return &types.SSHConnectionStatusResponse{
			Status:       "not_configured",
			Connected:    false,
			Message:      "SSH is not configured for this server",
			IsConfigured: false,
		}, nil
	}

	session, err := sshMgr.NewSessionWithRetry("")
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), serverID.String())
		return &types.SSHConnectionStatusResponse{
			Status:       "disconnected",
			Connected:    false,
			Message:      "Unable to connect to SSH server",
			IsConfigured: true,
		}, nil
	}
	session.Close()

	return &types.SSHConnectionStatusResponse{
		Status:       "connected",
		Connected:    true,
		Message:      "SSH connection is active",
		IsConfigured: true,
	}, nil
}
