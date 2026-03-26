package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *MCPService) AddServer(req *validation.CreateServerRequest, orgID, userID uuid.UUID) (*shared_types.MCPServer, error) {
	s.logger.Log(logger.Info, "Adding MCP server", req.Name)

	existing, err := s.storage.GetServerByName(orgID, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateName
	}

	var customURL *string
	if req.CustomURL != "" {
		customURL = &req.CustomURL
	}

	server := &shared_types.MCPServer{
		ID:          uuid.New(),
		OrgID:       orgID,
		ProviderID:  req.ProviderID,
		Name:        req.Name,
		Credentials: req.Credentials,
		CustomURL:   customURL,
		Enabled:     req.Enabled,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.storage.CreateServer(server); err != nil {
		return nil, err
	}
	return server, nil
}
