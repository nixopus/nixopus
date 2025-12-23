package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/feature-flags/storage"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestFeatureFlagStorage(t *testing.T) {
	t.Run("should create and get feature flag", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		featureStorage := &storage.FeatureFlagStorage{DB: setup.DB, Ctx: setup.Ctx}

		_, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		featureFlag := types.FeatureFlag{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			FeatureName:    string(types.FeatureDomain),
			IsEnabled:      true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err = featureStorage.CreateFeatureFlag(featureFlag)
		assert.NoError(t, err)

		flags, err := featureStorage.GetFeatureFlags(org.ID)
		assert.NoError(t, err)
		assert.Len(t, flags, 1)
		assert.Equal(t, featureFlag.FeatureName, flags[0].FeatureName)
		assert.Equal(t, featureFlag.IsEnabled, flags[0].IsEnabled)
	})

	t.Run("should update existing feature flag", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		featureStorage := &storage.FeatureFlagStorage{DB: setup.DB, Ctx: setup.Ctx}

		_, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		featureFlag := types.FeatureFlag{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			FeatureName:    string(types.FeatureDomain),
			IsEnabled:      true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err = featureStorage.CreateFeatureFlag(featureFlag)
		assert.NoError(t, err)

		err = featureStorage.UpdateFeatureFlag(org.ID, string(types.FeatureDomain), false)
		assert.NoError(t, err)

		isEnabled, err := featureStorage.IsFeatureEnabled(org.ID, string(types.FeatureDomain))
		assert.NoError(t, err)
		assert.False(t, isEnabled)
	})

	t.Run("should create new feature flag on update if not exists", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		featureStorage := &storage.FeatureFlagStorage{DB: setup.DB, Ctx: setup.Ctx}

		_, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		err = featureStorage.UpdateFeatureFlag(org.ID, string(types.FeatureDomain), true)
		assert.NoError(t, err)

		flags, err := featureStorage.GetFeatureFlags(org.ID)
		assert.NoError(t, err)
		assert.Len(t, flags, 1)
		assert.Equal(t, string(types.FeatureDomain), flags[0].FeatureName)
		assert.True(t, flags[0].IsEnabled)
	})

	t.Run("should handle transaction operations", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		featureStorage := &storage.FeatureFlagStorage{DB: setup.DB, Ctx: setup.Ctx}

		_, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		tx, err := featureStorage.BeginTx()
		assert.NoError(t, err)
		defer tx.Rollback()

		txStorage := featureStorage.WithTx(tx)

		err = txStorage.UpdateFeatureFlag(org.ID, string(types.FeatureDomain), true)
		assert.NoError(t, err)

		err = txStorage.UpdateFeatureFlag(org.ID, string(types.FeatureTerminal), false)
		assert.NoError(t, err)

		err = tx.Commit()
		assert.NoError(t, err)

		flags, err := featureStorage.GetFeatureFlags(org.ID)
		assert.NoError(t, err)
		assert.Len(t, flags, 2)

		domainEnabled, err := featureStorage.IsFeatureEnabled(org.ID, string(types.FeatureDomain))
		assert.NoError(t, err)
		assert.True(t, domainEnabled)

		terminalEnabled, err := featureStorage.IsFeatureEnabled(org.ID, string(types.FeatureTerminal))
		assert.NoError(t, err)
		assert.True(t, terminalEnabled)
	})

	t.Run("should rollback transaction on error", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		featureStorage := &storage.FeatureFlagStorage{DB: setup.DB, Ctx: setup.Ctx}

		_, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		tx, err := featureStorage.BeginTx()
		assert.NoError(t, err)
		defer tx.Rollback()

		txStorage := featureStorage.WithTx(tx)

		err = txStorage.UpdateFeatureFlag(org.ID, string(types.FeatureDomain), true)
		assert.NoError(t, err)

		err = txStorage.UpdateFeatureFlag(uuid.Nil, string(types.FeatureTerminal), false)
		assert.Error(t, err)

		// Transaction should be rolled back
		flags, err := featureStorage.GetFeatureFlags(org.ID)
		assert.NoError(t, err)
		assert.Empty(t, flags)
	})
}
