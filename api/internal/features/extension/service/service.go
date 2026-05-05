package service

import (
	"context"
	"fmt"

	"github.com/nixopus/nixopus/api/internal/cache"
	"github.com/nixopus/nixopus/api/internal/features/extension/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
)

// extCache abstracts Redis extension caching for tests and production (*cache.Cache).
type extCache interface {
	GetExtension(ctx context.Context, id string) (*types.Extension, error)
	GetExtensionByExtID(ctx context.Context, extensionID string) (*types.Extension, error)
	GetExtensionCategories(ctx context.Context) ([]types.ExtensionCategory, error)
	SetExtension(ctx context.Context, ext *types.Extension) error
	SetExtensionCategories(ctx context.Context, cats []types.ExtensionCategory) error
}

var _ extCache = (*cache.Cache)(nil)

type ExtensionService struct {
	storage storage.ExtensionStorageInterface
	cache   extCache
	ctx     context.Context
	logger  logger.Logger
}

func NewExtensionService(
	ctx context.Context,
	l logger.Logger,
	repository storage.ExtensionStorageInterface,
	appCache extCache,
) *ExtensionService {
	return &ExtensionService{
		storage: repository,
		cache:   appCache,
		ctx:     ctx,
		logger:  l,
	}
}

func (s *ExtensionService) GetExtension(id string) (*types.Extension, error) {
	if s.cache != nil {
		cached, err := s.cache.GetExtension(s.ctx, id)
		if err != nil {
			s.logger.Log(logger.Debug, fmt.Sprintf("extension service: GetExtension cache read failed, falling through to db: %v", err), id)
		}
		if cached != nil {
			return cached, nil
		}
	}

	extension, err := s.storage.GetExtension(id)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("extension service: GetExtension: %v", err), id)
		return nil, err
	}

	if s.cache != nil {
		if cacheErr := s.cache.SetExtension(s.ctx, extension); cacheErr != nil {
			s.logger.Log(logger.Warning, fmt.Sprintf("extension service: GetExtension failed to cache extension: %v", cacheErr), id)
		}
	}

	return extension, nil
}

func (s *ExtensionService) GetExtensionByID(extensionID string) (*types.Extension, error) {
	if s.cache != nil {
		cached, err := s.cache.GetExtensionByExtID(s.ctx, extensionID)
		if err != nil {
			s.logger.Log(logger.Debug, fmt.Sprintf("extension service: GetExtensionByID cache read failed, falling through to db: %v", err), extensionID)
		}
		if cached != nil {
			return cached, nil
		}
	}

	extension, err := s.storage.GetExtensionByID(extensionID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("extension service: GetExtensionByID: %v", err), extensionID)
		return nil, err
	}

	if s.cache != nil {
		if cacheErr := s.cache.SetExtension(s.ctx, extension); cacheErr != nil {
			s.logger.Log(logger.Warning, fmt.Sprintf("extension service: GetExtensionByID failed to cache extension: %v", cacheErr), extensionID)
		}
	}

	return extension, nil
}

func (s *ExtensionService) ListExtensions(params types.ExtensionListParams) (*types.ExtensionListResponse, error) {
	response, err := s.storage.ListExtensions(params)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("extension service: ListExtensions: %v", err), "")
		return nil, err
	}
	return response, nil
}

func (s *ExtensionService) ListCategories() ([]types.ExtensionCategory, error) {
	if s.cache != nil {
		cached, err := s.cache.GetExtensionCategories(s.ctx)
		if err != nil {
			s.logger.Log(logger.Debug, fmt.Sprintf("extension service: ListCategories cache read failed, falling through to db: %v", err), "")
		}
		if cached != nil {
			return cached, nil
		}
	}

	cats, err := s.storage.ListCategories()
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("extension service: ListCategories: %v", err), "")
		return nil, err
	}

	if s.cache != nil {
		if cacheErr := s.cache.SetExtensionCategories(s.ctx, cats); cacheErr != nil {
			s.logger.Log(logger.Warning, fmt.Sprintf("extension service: ListCategories failed to cache categories: %v", cacheErr), "")
		}
	}

	return cats, nil
}
