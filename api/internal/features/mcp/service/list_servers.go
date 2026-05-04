package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *MCPService) ListServers(orgID uuid.UUID, params storage.ListServersParams) ([]shared_types.MCPServer, int, error) {
	s.logger.Log(logger.Info, "mcp service: ListServers", fmt.Sprintf("org_id=%s enabled_only=%t page=%d limit=%d", orgID, params.EnabledOnly, params.Page, params.Limit))
	return s.storage.ListServers(orgID, params)
}
