package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/feature-flags/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type FeatureFlagService struct {
	storage storage.FeatureFlagRepository
	logger  logger.Logger
	ctx     context.Context
}

func NewFeatureFlagService(storage storage.FeatureFlagRepository, logger logger.Logger, ctx context.Context) *FeatureFlagService {
	return &FeatureFlagService{
		storage: storage,
		logger:  logger,
		ctx:     ctx,
	}
}

func (s *FeatureFlagService) GetFeatureFlags(organizationID uuid.UUID) ([]types.FeatureFlag, error) {
	s.logger.Log(logger.Info, "getting feature flags", "")
	flags, err := s.storage.GetFeatureFlags(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature flags: %w", err)
	}

	if len(flags) == 0 {
		s.logger.Log(logger.Info, "no feature flags found, creating defaults", "")

		tx, err := s.storage.BeginTx()
		if err != nil {
			s.logger.Log(logger.Error, "failed to start transaction", err.Error())
			return nil, fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		txStorage := s.storage.WithTx(tx)
		defaultFeatures := []types.FeatureName{types.FeatureDomain, types.FeatureTerminal, types.FeatureNotifications, types.FeatureFileManager, types.FeatureSelfHosted, types.FeatureAudit, types.FeatureGithubConnector, types.FeatureMonitoring, types.FeatureContainer}
		defaultFlags := make([]types.FeatureFlag, 0, len(defaultFeatures))

		for _, feature := range defaultFeatures {
			err := txStorage.UpdateFeatureFlag(organizationID, string(feature), true)
			if err != nil {
				s.logger.Log(logger.Error, fmt.Sprintf("failed to create default feature flag %s", feature), err.Error())
				return nil, fmt.Errorf("failed to create default feature flag %s: %w", feature, err)
			}
			defaultFlags = append(defaultFlags, types.FeatureFlag{
				OrganizationID: organizationID,
				FeatureName:    string(feature),
				IsEnabled:      true,
			})
		}

		if err := tx.Commit(); err != nil {
			s.logger.Log(logger.Error, "failed to commit transaction", err.Error())
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return defaultFlags, nil
	}

	return flags, nil
}

func (s *FeatureFlagService) UpdateFeatureFlag(organizationID uuid.UUID, req types.UpdateFeatureFlagRequest) error {
	s.logger.Log(logger.Info, "updating feature flag", "")
	return s.storage.UpdateFeatureFlag(organizationID, req.FeatureName, req.IsEnabled)
}

func (s *FeatureFlagService) IsFeatureEnabled(organizationID uuid.UUID, featureName string) (bool, error) {
	return s.storage.IsFeatureEnabled(organizationID, featureName)
}
