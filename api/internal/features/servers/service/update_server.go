package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/validation"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// UpdateServer updates an existing server
func (s *ServersService) UpdateServer(req types.UpdateServerRequest, userID string) (*shared_types.Server, error) {
	s.logger.Log(logger.Info, "update server request received", fmt.Sprintf("server_id=%s, user_id=%s", req.ID, userID))

	_, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return nil, types.ErrInvalidUserID
	}

	if userID == "" {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return nil, types.ErrInvalidUserID
	}

	validator := validation.NewValidator(s.storage)
	if err := validator.ValidateUpdateServerRequest(req); err != nil {
		return nil, err
	}

	tx, err := s.storage.BeginTx()
	if err != nil {
		s.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return nil, types.ErrFailedToUpdateServer
	}
	defer tx.Rollback()

	txStorage := s.storage.WithTx(tx)

	// Get the existing server
	existingServer, err := txStorage.GetServer(req.ID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving server", err.Error())
		return nil, types.ErrServerNotFound
	}

	if existingServer == nil {
		s.logger.Log(logger.Error, "server not found", fmt.Sprintf("server_id=%s", req.ID))
		return nil, types.ErrServerNotFound
	}

	// Check if the user owns the server
	if existingServer.UserID.String() != userID {
		s.logger.Log(logger.Error, "user does not own server", fmt.Sprintf("server_id=%s, user_id=%s", req.ID, userID))
		return nil, types.ErrPermissionDenied
	}

	// Check if another server with the same name exists (excluding current server)
	if req.Name != existingServer.Name {
		existingServerByName, err := txStorage.GetServerName(req.Name, existingServer.OrganizationID)
		if err != nil {
			s.logger.Log(logger.Debug, "error while checking existing server by name", err.Error())
		}

		if existingServerByName != nil && existingServerByName.ID != existingServer.ID {
			s.logger.Log(logger.Error, "server with name already exists", fmt.Sprintf("server_name=%s", req.Name))
			return nil, types.ErrServerAlreadyExists
		}
	}

	// Check if another server with the same host:port exists (excluding current server)
	if req.Host != existingServer.Host || req.Port != existingServer.Port {
		existingServerByHost, err := txStorage.GetServerByHost(req.Host, req.Port, existingServer.OrganizationID)
		if err != nil {
			s.logger.Log(logger.Debug, "error while checking existing server by host", err.Error())
		}

		if existingServerByHost != nil && existingServerByHost.ID != existingServer.ID {
			s.logger.Log(logger.Error, "server with host already exists", fmt.Sprintf("host=%s, port=%d", req.Host, req.Port))
			return nil, types.ErrServerHostAlreadyExists
		}
	}

	// Update the server - preserve existing auth credentials if not provided
	sshPassword := existingServer.SSHPassword
	if req.SSHPassword != nil {
		sshPassword = req.SSHPassword
	}

	sshPrivateKeyPath := existingServer.SSHPrivateKeyPath
	if req.SSHPrivateKeyPath != nil {
		sshPrivateKeyPath = req.SSHPrivateKeyPath
	}

	updatedServer := &shared_types.Server{
		ID:                existingServer.ID,
		Name:              req.Name,
		Description:       req.Description,
		Host:              req.Host,
		Port:              req.Port,
		Username:          req.Username,
		SSHPassword:       sshPassword,
		SSHPrivateKeyPath: sshPrivateKeyPath,
		Status:            existingServer.Status,
		CreatedAt:         existingServer.CreatedAt,
		UpdatedAt:         time.Now(),
		DeletedAt:         existingServer.DeletedAt,
		UserID:            existingServer.UserID,
		OrganizationID:    existingServer.OrganizationID,
	}

	if err := txStorage.UpdateServer(updatedServer); err != nil {
		s.logger.Log(logger.Error, "error while updating server", err.Error())
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return nil, types.ErrFailedToUpdateServer
	}

	return updatedServer, nil
}
