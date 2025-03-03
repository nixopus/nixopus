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
	l       logger.Logger
	storage storage.GithubConnectorStorage
}

func NewGithubConnectorService(store *shared_storage.Store, ctx context.Context, l logger.Logger) *GithubConnectorService {
	return &GithubConnectorService{
		store: store,
		ctx:   ctx,
		l:     l,
		storage: storage.GithubConnectorStorage{
			DB:  store.DB,
			Ctx: ctx,
		},
	}
}
