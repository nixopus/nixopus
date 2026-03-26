package service

import (
	"context"
	"errors"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/storage"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
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
