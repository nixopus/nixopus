package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/feature-flags/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
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
	orgStr := organizationID.String()
	s.logger.Log(logger.Info, "feature flags service: GetFeatureFlags", fmt.Sprintf("org_id=%s", orgStr))
	flags, err := s.storage.GetFeatureFlags(organizationID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("feature flags service: GetFeatureFlags: %v", err), fmt.Sprintf("org_id=%s", orgStr))
		return nil, fmt.Errorf("failed to get feature flags: %w", err)
	}

	if len(flags) == 0 {
		s.logger.Log(logger.Info, "feature flags service: GetFeatureFlags seeding defaults", fmt.Sprintf("org_id=%s", orgStr))

		tx, err := s.storage.BeginTx()
		if err != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("feature flags service: GetFeatureFlags BeginTx: %v", err), fmt.Sprintf("org_id=%s", orgStr))
			return nil, fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		txStorage := s.storage.WithTx(tx)
		defaultFeatures := []types.FeatureName{types.FeatureDomain, types.FeatureTerminal, types.FeatureNotifications, types.FeatureSelfHosted, types.FeatureAudit, types.FeatureGithubConnector, types.FeatureMonitoring, types.FeatureContainer, types.FeatureMCP}
		defaultFlags := make([]types.FeatureFlag, 0, len(defaultFeatures))

		for _, feature := range defaultFeatures {
			err := txStorage.UpdateFeatureFlag(organizationID, string(feature), true)
			if err != nil {
				s.logger.Log(logger.Error, fmt.Sprintf("feature flags service: GetFeatureFlags default flag %s: %v", feature, err), fmt.Sprintf("org_id=%s", orgStr))
				return nil, fmt.Errorf("failed to create default feature flag %s: %w", feature, err)
			}
			defaultFlags = append(defaultFlags, types.FeatureFlag{
				OrganizationID: organizationID,
				FeatureName:    string(feature),
				IsEnabled:      true,
			})
		}

		if err := tx.Commit(); err != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("feature flags service: GetFeatureFlags Commit: %v", err), fmt.Sprintf("org_id=%s", orgStr))
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		s.logger.Log(logger.Info, "feature flags service: GetFeatureFlags defaults committed", fmt.Sprintf("org_id=%s count=%d", orgStr, len(defaultFlags)))
		return defaultFlags, nil
	}

	s.logger.Log(logger.Info, "feature flags service: GetFeatureFlags ok", fmt.Sprintf("org_id=%s count=%d", orgStr, len(flags)))
	return flags, nil
}

func (s *FeatureFlagService) UpdateFeatureFlag(organizationID uuid.UUID, req types.UpdateFeatureFlagRequest) error {
	ctxStr := fmt.Sprintf("org_id=%s feature_name=%s is_enabled=%t", organizationID, req.FeatureName, req.IsEnabled)
	s.logger.Log(logger.Info, "feature flags service: UpdateFeatureFlag", ctxStr)
	if err := s.storage.UpdateFeatureFlag(organizationID, req.FeatureName, req.IsEnabled); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("feature flags service: UpdateFeatureFlag: %v", err), ctxStr)
		return err
	}
	s.logger.Log(logger.Info, "feature flags service: UpdateFeatureFlag ok", ctxStr)
	return nil
}

func (s *FeatureFlagService) IsFeatureEnabled(organizationID uuid.UUID, featureName string) (bool, error) {
	ctxStr := fmt.Sprintf("org_id=%s feature_name=%s", organizationID, featureName)
	enabled, err := s.storage.IsFeatureEnabled(organizationID, featureName)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("feature flags service: IsFeatureEnabled: %v", err), ctxStr)
		return false, err
	}
	s.logger.Log(logger.Debug, "feature flags service: IsFeatureEnabled ok", fmt.Sprintf("%s enabled=%t", ctxStr, enabled))
	return enabled, nil
}
