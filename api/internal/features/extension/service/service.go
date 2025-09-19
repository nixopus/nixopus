package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type ExtensionService struct {
	store   *shared_storage.Store
	storage storage.ExtensionStorageInterface
	ctx     context.Context
	logger  logger.Logger
}

func NewExtensionService(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	storage storage.ExtensionStorageInterface,
) *ExtensionService {
	return &ExtensionService{
		store:   store,
		storage: storage,
		ctx:     ctx,
		logger:  l,
	}
}

func (s *ExtensionService) CreateExtension(extension *types.Extension) error {
	if err := s.storage.CreateExtension(extension); err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}
	return nil
}

func (s *ExtensionService) GetExtension(id string) (*types.Extension, error) {
	extension, err := s.storage.GetExtension(id)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return extension, nil
}

func (s *ExtensionService) GetExtensionByID(extensionID string) (*types.Extension, error) {
	extension, err := s.storage.GetExtensionByID(extensionID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return extension, nil
}

func (s *ExtensionService) UpdateExtension(extension *types.Extension) error {
	if err := s.storage.UpdateExtension(extension); err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}
	return nil
}

func (s *ExtensionService) DeleteExtension(id string) error {
	if err := s.storage.DeleteExtension(id); err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}
	return nil
}

func (s *ExtensionService) ListExtensions(params types.ExtensionListParams) (*types.ExtensionListResponse, error) {
	response, err := s.storage.ListExtensions(params)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return response, nil
}
