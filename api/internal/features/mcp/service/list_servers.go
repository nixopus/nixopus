package service

import (
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *MCPService) ListServers(orgID uuid.UUID, enabledOnly bool) ([]shared_types.MCPServer, error) {
	s.logger.Log(logger.Info, "Listing MCP servers", orgID.String())
	return s.storage.ListServers(orgID, enabledOnly)
}
