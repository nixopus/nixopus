package service

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/extension/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
)

type ExtensionService struct {
	store   *shared_storage.Store
	storage storage.ExtensionStorageInterface
	cache   *ExtensionCache
	ctx     context.Context
	logger  logger.Logger
}

func NewExtensionService(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	storage storage.ExtensionStorageInterface,
	redisURL string,
) *ExtensionService {
	var extCache *ExtensionCache
	if redisURL != "" {
		c, err := NewExtensionCache(redisURL)
		if err != nil {
			l.Log(logger.Error, "failed to create extension cache, proceeding without cache", err.Error())
		} else {
			extCache = c
		}
	}

	return &ExtensionService{
		store:   store,
		storage: storage,
		cache:   extCache,
		ctx:     ctx,
		logger:  l,
	}
}

func (s *ExtensionService) GetExtension(id string) (*types.Extension, error) {
	if s.cache != nil {
		cached, err := s.cache.GetExtension(s.ctx, id)
		if err != nil {
			s.logger.Log(logger.Error, "extension cache read failed, falling through to db", err.Error())
		}
		if cached != nil {
			return cached, nil
		}
	}

	extension, err := s.storage.GetExtension(id)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	if s.cache != nil {
		if cacheErr := s.cache.SetExtension(s.ctx, extension); cacheErr != nil {
			s.logger.Log(logger.Error, "failed to cache extension", cacheErr.Error())
		}
	}

	return extension, nil
}

func (s *ExtensionService) GetExtensionByID(extensionID string) (*types.Extension, error) {
	if s.cache != nil {
		cached, err := s.cache.GetExtensionByExtID(s.ctx, extensionID)
		if err != nil {
			s.logger.Log(logger.Error, "extension cache read failed, falling through to db", err.Error())
		}
		if cached != nil {
			return cached, nil
		}
	}

	extension, err := s.storage.GetExtensionByID(extensionID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	if s.cache != nil {
		if cacheErr := s.cache.SetExtension(s.ctx, extension); cacheErr != nil {
			s.logger.Log(logger.Error, "failed to cache extension", cacheErr.Error())
		}
	}

	return extension, nil
}

func (s *ExtensionService) ListExtensions(params types.ExtensionListParams) (*types.ExtensionListResponse, error) {
	response, err := s.storage.ListExtensions(params)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return response, nil
}

func (s *ExtensionService) ListCategories() ([]types.ExtensionCategory, error) {
	if s.cache != nil {
		cached, err := s.cache.GetCategories(s.ctx)
		if err != nil {
			s.logger.Log(logger.Error, "categories cache read failed, falling through to db", err.Error())
		}
		if cached != nil {
			return cached, nil
		}
	}

	cats, err := s.storage.ListCategories()
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	if s.cache != nil {
		if cacheErr := s.cache.SetCategories(s.ctx, cats); cacheErr != nil {
			s.logger.Log(logger.Error, "failed to cache categories", cacheErr.Error())
		}
	}

	return cats, nil
}

func (s *ExtensionService) invalidateExtensionCache(id string, extensionID string) {
	if s.cache == nil {
		return
	}
	if err := s.cache.InvalidateExtension(s.ctx, id, extensionID); err != nil {
		s.logger.Log(logger.Error, "failed to invalidate extension cache", err.Error())
	}
}

func (s *ExtensionService) invalidateCategoriesCache() {
	if s.cache == nil {
		return
	}
	if err := s.cache.InvalidateCategories(s.ctx); err != nil {
		s.logger.Log(logger.Error, "failed to invalidate categories cache", err.Error())
	}
}
