package service

import (
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (s *MCPService) DeleteServer(id string, orgID uuid.UUID) error {
	s.logger.Log(logger.Info, "Deleting MCP server", id)

	serverID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	server, err := s.storage.GetServerByID(serverID)
	if err != nil {
		return err
	}
	if server == nil || server.OrgID != orgID {
		return ErrServerNotFound
	}

	return s.storage.DeleteServer(serverID)
}
