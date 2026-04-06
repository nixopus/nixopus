package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/storage"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

var ErrServerNotFound = errors.New("MCP server not found")
var ErrDuplicateName = errors.New("an MCP server with this name already exists")

type MCPService struct {
	storage storage.MCPRepository
	ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewMCPService(store *shared_storage.Store, ctx context.Context, l logger.Logger, repo storage.MCPRepository) *MCPService {
	return &MCPService{storage: repo, ctx: ctx, store: store, logger: l}
}

func (s *MCPService) GetServerByID(id, orgID uuid.UUID) (*shared_types.MCPServer, error) {
	server, err := s.storage.GetServerByID(id)
	if err != nil {
		return nil, err
	}
	if server == nil || server.OrgID != orgID {
		return nil, ErrServerNotFound
	}
	return server, nil
}
