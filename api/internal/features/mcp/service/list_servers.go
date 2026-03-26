package service

import (
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *MCPService) ListServers(orgID uuid.UUID, params storage.ListServersParams) ([]shared_types.MCPServer, int, error) {
	s.logger.Log(logger.Info, "Listing MCP servers", orgID.String())
	return s.storage.ListServers(orgID, params)
}
