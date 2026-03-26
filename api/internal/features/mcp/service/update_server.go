package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *MCPService) UpdateServer(req *validation.UpdateServerRequest, orgID uuid.UUID) (*shared_types.MCPServer, error) {
	s.logger.Log(logger.Info, "Updating MCP server", req.ID)

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}

	server, err := s.storage.GetServerByID(id)
	if err != nil {
		return nil, err
	}
	if server == nil || server.OrgID != orgID {
		return nil, ErrServerNotFound
	}

	server.Name = req.Name
	server.Credentials = req.Credentials
	server.Enabled = req.Enabled
	server.UpdatedAt = time.Now()
	if req.CustomURL != "" {
		server.CustomURL = &req.CustomURL
	} else {
		server.CustomURL = nil
	}

	if err := s.storage.UpdateServer(server); err != nil {
		return nil, err
	}
	return server, nil
}
