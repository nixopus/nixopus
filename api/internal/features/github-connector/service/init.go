package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type GithubConnectorService struct {
	store   *shared_storage.Store
	ctx     context.Context
	logger       logger.Logger
	storage storage.GithubConnectorRepository
}

func NewGithubConnectorService(store *shared_storage.Store, ctx context.Context, l logger.Logger, GithubConnectorRepository storage.GithubConnectorRepository) *GithubConnectorService {
	return &GithubConnectorService{
		store:   store,
		ctx:     ctx,
		logger:       l,
		storage: GithubConnectorRepository,
	}
}
