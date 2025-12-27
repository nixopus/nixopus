package routes

import (
	"github.com/go-fuego/fuego"
	feature_flags_controller "github.com/raghavyuva/nixopus-api/internal/features/feature-flags/controller"
)

// RegisterFeatureFlagRoutes registers feature flag routes
func (router *Router) RegisterFeatureFlagRoutes(readGroup *fuego.Server, writeGroup *fuego.Server, featureFlagController *feature_flags_controller.FeatureFlagController) {
	fuego.Get(readGroup, "", featureFlagController.GetFeatureFlags)
	fuego.Put(writeGroup, "", featureFlagController.UpdateFeatureFlag)
	fuego.Get(readGroup, "/check", featureFlagController.IsFeatureEnabled)
}
