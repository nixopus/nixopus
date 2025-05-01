package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/cache"
	feature_flags_storage "github.com/raghavyuva/nixopus-api/internal/features/feature-flags/storage"
	appStorage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func FeatureFlagMiddleware(next http.Handler, app *appStorage.App, featureName string, cache *cache.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, ok := r.Context().Value(types.OrganizationIDKey).(string)
		if !ok {
			utils.SendErrorResponse(w, "Organization ID not found in context", http.StatusBadRequest)
			return
		}

		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			utils.SendErrorResponse(w, "Invalid organization ID", http.StatusBadRequest)
			return
		}

		disableCache := r.Header.Get("X-Disable-Cache")
		if disableCache == "true" {
			cache = nil
		}

		featureFlagStorage := &feature_flags_storage.FeatureFlagStorage{DB: app.Store.DB, Ctx: app.Ctx}

		if cache != nil {
			enabled, err := cache.GetFeatureFlag(app.Ctx, organizationID.String(), featureName)
			if err == nil {
				if enabled {
					next.ServeHTTP(w, r)
					return
				}
				utils.SendErrorResponse(w, "Feature is disabled", http.StatusForbidden)
				return
			}
		}

		isEnabled, err := featureFlagStorage.IsFeatureEnabled(organizationID, featureName)
		if err != nil {
			utils.SendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !isEnabled {
			utils.SendErrorResponse(w, "Feature is disabled", http.StatusForbidden)
			return
		}

		if cache != nil {
			cache.SetFeatureFlag(app.Ctx, organizationID.String(), featureName, isEnabled)
		}

		next.ServeHTTP(w, r)
	})
}
