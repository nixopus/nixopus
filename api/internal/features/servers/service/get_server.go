package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetServer retrieves a server by ID for a specific user
func (s *ServersService) GetServer(serverID string, userID string) (*shared_types.Server, error) {
	s.logger.Log(logger.Info, "get server request received", fmt.Sprintf("server_id=%s, user_id=%s", serverID, userID))

	_, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return nil, types.ErrInvalidUserID
	}

	if userID == "" {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return nil, types.ErrInvalidUserID
	}

	_, err = uuid.Parse(serverID)
	if err != nil {
		s.logger.Log(logger.Error, "invalid server id", fmt.Sprintf("server_id=%s", serverID))
		return nil, types.ErrInvalidServerID
	}

	server, err := s.storage.GetServer(serverID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving server", err.Error())
		return nil, types.ErrServerNotFound
	}

	if server == nil {
		s.logger.Log(logger.Error, "server not found", fmt.Sprintf("server_id=%s", serverID))
		return nil, types.ErrServerNotFound
	}

	if server.UserID.String() != userID {
		s.logger.Log(logger.Error, "user does not own server", fmt.Sprintf("server_id=%s, user_id=%s, owner_id=%s", serverID, userID, server.UserID.String()))
		return nil, types.ErrPermissionDenied
	}

	return server, nil
}
