package tests

import (
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/feature-flags/service"
	"github.com/raghavyuva/nixopus-api/internal/features/feature-flags/storage"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestFeatureFlags(t *testing.T) {
	t.Run("should get default feature flags when none exist", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &storage.FeatureFlagStorage{DB: setup.DB, Ctx: setup.Ctx}
		featureService := service.NewFeatureFlagService(storage, setup.Logger, setup.Ctx)

		_, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		flags, err := featureService.GetFeatureFlags(org.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, flags)

		defaultFeatures := []types.FeatureName{
			types.FeatureDomain,
			types.FeatureTerminal,
			types.FeatureNotifications,
			types.FeatureFileManager,
			types.FeatureSelfHosted,
			types.FeatureAudit,
			types.FeatureGithubConnector,
			types.FeatureMonitoring,
			types.FeatureContainer,
		}

		for _, feature := range defaultFeatures {
			found := false
			for _, flag := range flags {
				if flag.FeatureName == string(feature) {
					assert.True(t, flag.IsEnabled)
					found = true
					break
				}
			}
			assert.True(t, found, "Default feature %s not found", feature)
		}
	})

	t.Run("should update feature flag", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &storage.FeatureFlagStorage{DB: setup.DB, Ctx: setup.Ctx}
		featureService := service.NewFeatureFlagService(storage, setup.Logger, setup.Ctx)

		_, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		_, err = featureService.GetFeatureFlags(org.ID)
		assert.NoError(t, err)

		req := types.UpdateFeatureFlagRequest{
			FeatureName: string(types.FeatureDomain),
			IsEnabled:   false,
		}

		err = featureService.UpdateFeatureFlag(org.ID, req)
		assert.NoError(t, err)

		isEnabled, err := featureService.IsFeatureEnabled(org.ID, string(types.FeatureDomain))
		assert.NoError(t, err)
		assert.False(t, isEnabled)
	})

	t.Run("should return true for non-existent feature flag", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &storage.FeatureFlagStorage{DB: setup.DB, Ctx: setup.Ctx}
		featureService := service.NewFeatureFlagService(storage, setup.Logger, setup.Ctx)

		_, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		isEnabled, err := featureService.IsFeatureEnabled(org.ID, "non_existent_feature")
		assert.NoError(t, err)
		assert.True(t, isEnabled)
	})

	t.Run("should handle multiple feature flag updates", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &storage.FeatureFlagStorage{DB: setup.DB, Ctx: setup.Ctx}
		featureService := service.NewFeatureFlagService(storage, setup.Logger, setup.Ctx)

		_, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		_, err = featureService.GetFeatureFlags(org.ID)
		assert.NoError(t, err)

		updates := []struct {
			feature types.FeatureName
			enabled bool
		}{
			{types.FeatureDomain, false},
			{types.FeatureTerminal, true},
			{types.FeatureNotifications, false},
		}

		for _, update := range updates {
			req := types.UpdateFeatureFlagRequest{
				FeatureName: string(update.feature),
				IsEnabled:   update.enabled,
			}
			err = featureService.UpdateFeatureFlag(org.ID, req)
			assert.NoError(t, err)
		}

		for _, update := range updates {
			isEnabled, err := featureService.IsFeatureEnabled(org.ID, string(update.feature))
			assert.NoError(t, err)
			assert.Equal(t, update.enabled, isEnabled)
		}
	})
}
