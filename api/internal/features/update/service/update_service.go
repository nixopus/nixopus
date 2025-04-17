package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/update/types"
	"github.com/raghavyuva/nixopus-api/internal/storage"
)

type UpdateService struct {
	storage *storage.App
	logger  *logger.Logger
	ctx     context.Context
}

func NewUpdateService(storage *storage.App, logger *logger.Logger, ctx context.Context) *UpdateService {
	return &UpdateService{
		storage: storage,
		logger:  logger,
		ctx:     ctx,
	}
}

// TODO: Implement update check service
func (s *UpdateService) CheckForUpdates() (*types.UpdateCheckResponse, error) {
	return &types.UpdateCheckResponse{
		UpdateAvailable: true,
		LastChecked:     time.Now(),
	}, nil
}

// TODO: Implement update service
func (s *UpdateService) PerformUpdate() error {
	return nil
}

func (s *UpdateService) GetUserAutoUpdatePreference(userID uuid.UUID) (bool, error) {
	var autoUpdate bool

	err := s.storage.Store.DB.NewSelect().
		TableExpr("user_settings").
		Column("auto_update").
		Where("user_id = ?", userID).
		Scan(s.ctx, &autoUpdate)

	if err != nil {
		return false, fmt.Errorf("failed to get user settings: %w", err)
	}

	return autoUpdate, nil
}
