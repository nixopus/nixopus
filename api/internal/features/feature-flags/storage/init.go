package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type FeatureFlagStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
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

func (s *FeatureFlagStorage) getDB() bun.IDB {
	if s.tx != nil {
		return s.tx
	}
	return s.DB
}

func (s *FeatureFlagStorage) GetFeatureFlags(organizationID uuid.UUID) ([]types.FeatureFlag, error) {
	var flags []types.FeatureFlag
	err := s.getDB().NewSelect().
		Model(&flags).
		Where("organization_id = ?", organizationID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get feature flags: %w", err)
	}

	return flags, nil
}

func (s *FeatureFlagStorage) UpdateFeatureFlag(organizationID uuid.UUID, featureName string, isEnabled bool) error {
	flag := &types.FeatureFlag{}
	err := s.getDB().NewSelect().
		Model(flag).
		Where("organization_id = ?", organizationID).
		Where("feature_name = ?", featureName).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			flag = &types.FeatureFlag{
				ID:             uuid.New(),
				OrganizationID: organizationID,
				FeatureName:    featureName,
				IsEnabled:      isEnabled,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			_, err = s.getDB().NewInsert().Model(flag).Exec(s.Ctx)
			return err
		}
		return fmt.Errorf("failed to get feature flag: %w", err)
	}

	flag.IsEnabled = isEnabled
	flag.UpdatedAt = time.Now()
	_, err = s.getDB().NewUpdate().
		Model(flag).
		Where("id = ?", flag.ID).
		Exec(s.Ctx)

	return err
}

func (s *FeatureFlagStorage) IsFeatureEnabled(organizationID uuid.UUID, featureName string) (bool, error) {
	var isEnabled bool
	err := s.getDB().NewSelect().
		TableExpr("feature_flags").
		Column("is_enabled").
		Where("organization_id = ?", organizationID).
		Where("feature_name = ?", featureName).
		Where("deleted_at IS NULL").
		Scan(s.Ctx, &isEnabled)

	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil // Default to enabled if not configured
		}
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
		DB:  s.DB,
		Ctx: s.Ctx,
		tx:  &tx,
	}
}
