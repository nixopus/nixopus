package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type FeatureFlagStorage struct {
	DB     *bun.DB
	Ctx    context.Context
	tx     *bun.Tx
	Logger *logger.Logger // optional; nil disables storage logs
}

type FeatureFlagRepository interface {
	GetFeatureFlags(organizationID uuid.UUID) ([]types.FeatureFlag, error)
	UpdateFeatureFlag(organizationID uuid.UUID, featureName string, isEnabled bool) error
	IsFeatureEnabled(organizationID uuid.UUID, featureName string) (bool, error)
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) FeatureFlagRepository
}

func NewFeatureFlagStorage(db *bun.DB, ctx context.Context) *FeatureFlagStorage {
	return &FeatureFlagStorage{
		DB:  db,
		Ctx: ctx,
	}
}

func (s *FeatureFlagStorage) storageLog(sev logger.Severity, msg, data string) {
	if s.Logger == nil {
		return
	}
	s.Logger.Log(sev, msg, data)
}

func (s *FeatureFlagStorage) getDB() bun.IDB {
	if s.tx != nil {
		return s.tx
	}
	return s.DB
}

func (s *FeatureFlagStorage) GetFeatureFlags(organizationID uuid.UUID) ([]types.FeatureFlag, error) {
	ctxStr := fmt.Sprintf("org_id=%s", organizationID)
	s.storageLog(logger.Debug, "storage: GetFeatureFlags", ctxStr)
	var flags []types.FeatureFlag
	err := s.getDB().NewSelect().
		Model(&flags).
		Where("organization_id = ?", organizationID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)

	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetFeatureFlags: %v", err), ctxStr)
		return nil, fmt.Errorf("failed to get feature flags: %w", err)
	}

	s.storageLog(logger.Debug, "storage: GetFeatureFlags ok", fmt.Sprintf("%s count=%d", ctxStr, len(flags)))
	return flags, nil
}

func (s *FeatureFlagStorage) UpdateFeatureFlag(organizationID uuid.UUID, featureName string, isEnabled bool) error {
	ctxStr := fmt.Sprintf("org_id=%s feature_name=%s is_enabled=%t", organizationID, featureName, isEnabled)
	s.storageLog(logger.Debug, "storage: UpdateFeatureFlag", ctxStr)
	flag := &types.FeatureFlag{}
	err := s.getDB().NewSelect().
		Model(flag).
		Where("organization_id = ?", organizationID).
		Where("feature_name = ?", featureName).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			flag = &types.FeatureFlag{
				ID:             uuid.New(),
				OrganizationID: organizationID,
				FeatureName:    featureName,
				IsEnabled:      isEnabled,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			_, err = s.getDB().NewInsert().Model(flag).Exec(s.Ctx)
			if err != nil {
				s.storageLog(logger.Error, fmt.Sprintf("storage: UpdateFeatureFlag insert: %v", err), ctxStr)
			}
			return err
		}
		s.storageLog(logger.Error, fmt.Sprintf("storage: UpdateFeatureFlag select: %v", err), ctxStr)
		return fmt.Errorf("failed to get feature flag: %w", err)
	}

	flag.IsEnabled = isEnabled
	flag.UpdatedAt = time.Now()
	_, err = s.getDB().NewUpdate().
		Model(flag).
		Where("id = ?", flag.ID).
		Exec(s.Ctx)

	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: UpdateFeatureFlag update: %v", err), ctxStr)
	}
	return err
}

func (s *FeatureFlagStorage) IsFeatureEnabled(organizationID uuid.UUID, featureName string) (bool, error) {
	ctxStr := fmt.Sprintf("org_id=%s feature_name=%s", organizationID, featureName)
	s.storageLog(logger.Debug, "storage: IsFeatureEnabled", ctxStr)
	var isEnabled bool
	err := s.getDB().NewSelect().
		TableExpr("feature_flags").
		Column("is_enabled").
		Where("organization_id = ?", organizationID).
		Where("feature_name = ?", featureName).
		Where("deleted_at IS NULL").
		Scan(s.Ctx, &isEnabled)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.storageLog(logger.Debug, "storage: IsFeatureEnabled not found default enabled", ctxStr)
			return true, nil // Default to enabled if not configured
		}
		s.storageLog(logger.Error, fmt.Sprintf("storage: IsFeatureEnabled: %v", err), ctxStr)
		return false, fmt.Errorf("failed to check feature flag: %w", err)
	}

	return isEnabled, nil
}

func (s *FeatureFlagStorage) CreateFeatureFlag(featureFlag types.FeatureFlag) error {
	_, err := s.getDB().NewInsert().Model(&featureFlag).Exec(s.Ctx)
	return err
}

func (s *FeatureFlagStorage) BeginTx() (bun.Tx, error) {
	return s.DB.BeginTx(s.Ctx, nil)
}

func (s *FeatureFlagStorage) WithTx(tx bun.Tx) FeatureFlagRepository {
	return &FeatureFlagStorage{
		DB:     s.DB,
		Ctx:    s.Ctx,
		tx:     &tx,
		Logger: s.Logger,
	}
}
