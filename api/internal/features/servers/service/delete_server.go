package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/validation"
)

// DeleteServer deletes a server (soft delete)
func (s *ServersService) DeleteServer(req types.DeleteServerRequest, userID string) error {
	s.logger.Log(logger.Info, "delete server request received", fmt.Sprintf("server_id=%s, user_id=%s", req.ID, userID))

	_, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return types.ErrInvalidUserID
	}

	if userID == "" {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return types.ErrInvalidUserID
	}

	validator := validation.NewValidator(s.storage)
	if err := validator.ValidateDeleteServerRequest(req); err != nil {
		return err
	}

	tx, err := s.storage.BeginTx()
	if err != nil {
		s.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return types.ErrFailedToDeleteServer
	}
	defer tx.Rollback()

	txStorage := s.storage.WithTx(tx)

	// Get the existing server
	existingServer, err := txStorage.GetServer(req.ID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving server", err.Error())
		return types.ErrServerNotFound
	}

	if existingServer == nil {
		s.logger.Log(logger.Error, "server not found", fmt.Sprintf("server_id=%s", req.ID))
		return types.ErrServerNotFound
	}

	// Check if the user owns the server
	if existingServer.UserID.String() != userID {
		s.logger.Log(logger.Error, "user does not own server", fmt.Sprintf("server_id=%s, user_id=%s", req.ID, userID))
		return types.ErrPermissionDenied
	}

	// Soft delete the server
	if err := txStorage.DeleteServer(existingServer); err != nil {
		s.logger.Log(logger.Error, "error while deleting server", err.Error())
		return err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return types.ErrFailedToDeleteServer
	}

	s.logger.Log(logger.Info, "server deleted successfully", fmt.Sprintf("server_id=%s", req.ID))
	return nil
}
