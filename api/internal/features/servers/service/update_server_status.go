package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/validation"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *ServersService) UpdateServerStatus(req types.UpdateServerStatusRequest, userID string) (*shared_types.Server, error) {
	s.logger.Log(logger.Info, "update server status request received", fmt.Sprintf("server_id=%s, status=%s, user_id=%s", req.ID, req.Status, userID))

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
	if err := validator.ValidateUpdateServerStatusRequest(req); err != nil {
		return nil, err
	}

	tx, err := s.storage.BeginTx()
	if err != nil {
		s.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return nil, types.ErrFailedToUpdateServer
	}
	defer tx.Rollback()

	txStorage := s.storage.WithTx(tx)

	existingServer, err := txStorage.GetServer(req.ID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving server", err.Error())
		return nil, types.ErrServerNotFound
	}

	if existingServer == nil {
		s.logger.Log(logger.Error, "server not found", fmt.Sprintf("server_id=%s", req.ID))
		return nil, types.ErrServerNotFound
	}

	if existingServer.UserID.String() != userID {
		s.logger.Log(logger.Error, "user does not own server", fmt.Sprintf("server_id=%s, user_id=%s", req.ID, userID))
		return nil, types.ErrPermissionDenied
	}

	err = txStorage.UpdateServerStatus(req.ID, req.Status)
	if err != nil {
		s.logger.Log(logger.Error, "failed to update server status", err.Error())
		return nil, types.ErrFailedToUpdateServer
	}

	if err := tx.Commit(); err != nil {
		s.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return nil, types.ErrFailedToUpdateServer
	}

	updatedServer, err := s.storage.GetServer(req.ID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving updated server", err.Error())
		return nil, types.ErrServerNotFound
	}

	s.logger.Log(logger.Info, "server status updated successfully", fmt.Sprintf("server_id=%s, new_status=%s", req.ID, req.Status))

	return updatedServer, nil
}
