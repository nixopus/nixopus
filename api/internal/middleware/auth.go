package middleware

import (
	"context"
	"github.com/raghavyuva/nixopus-api/internal/cache"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"net/http"
	"strings"
)

// AuthMiddleware is a middleware that checks if the request has a valid
// SuperTokens session. If the session is valid, it adds both the user and
// the authenticated client to the request context.
func AuthMiddleware(next http.Handler, app *storage.App, cache *cache.Cache) http.Handler {
	return session.VerifySession(nil, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get the session from the request context
		sessionContainer := session.GetSessionFromRequestContext(ctx)
		userID := sessionContainer.GetUserID()

		// Optionally cache can be disabled by setting the X-Disable-Cache header to true
		disableCache := r.Header.Get("X-Disable-Cache")
		if disableCache == "true" {
			cache = nil
		}

		// Get user details from database
		userStorage := user_storage.UserStorage{
			DB:  app.Store.DB,
			Ctx: ctx,
		}

		var user *types.User
		var err error

		// Try to get user from cache first
		if cache != nil {
			if cachedUser, err := cache.GetUser(ctx, userID); err == nil && cachedUser != nil {
				user = cachedUser
			}
		}

		// If not in cache, fetch from database using SuperTokens user ID
		if user == nil {
			user, err = userStorage.FindUserBySupertokensID(userID)
			if err != nil {
				utils.SendErrorResponse(w, "User not found", http.StatusUnauthorized)
				return
			}

			// Cache the user for future requests
			if cache != nil {
				cache.SetUser(ctx, user.Email, user)
			}
		}

		// TODO: Add 2FA verification logic here (claims has to be sest by overriding supertokens session claims)
		ctx = context.WithValue(ctx, types.UserContextKey, user)

		if !isAuthEndpoint(r.URL.Path) {
			organizationID := r.Header.Get("X-Organization-Id")
			if organizationID == "" {
				utils.SendErrorResponse(w, "No organization ID provided", http.StatusBadRequest)
				return
			}

			var belongsToOrg bool
			if cache != nil {
				if cached, err := cache.GetOrgMembership(ctx, user.ID.String(), organizationID); err == nil {
					belongsToOrg = cached
				}
			}

			if !belongsToOrg {
				belongsToOrg, err = userStorage.UserBelongsToOrganization(user.ID.String(), organizationID)
				if err != nil {
					utils.SendErrorResponse(w, "Error verifying organization membership", http.StatusInternalServerError)
					return
				}

				if !belongsToOrg {
					utils.SendErrorResponse(w, "User does not belong to the specified organization", http.StatusForbidden)
					return
				}

				if cache != nil {
					cache.SetOrgMembership(ctx, user.ID.String(), organizationID, belongsToOrg)
				}
			}

			ctx = context.WithValue(ctx, types.OrganizationIDKey, organizationID)
		}

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/2fa-login",
		"/api/v1/auth/verify-2fa",
		"/api/v1/auth/refresh-token",
		"/api/v1/auth/logout",
		"/api/v1/auth/setup-2fa",
		"/api/v1/auth/disable-2fa",
		"/api/v1/auth/verify-email",
		"/api/v1/auth/send-verification-email",
		"/api/v1/auth/reset-password",
		"/api/v1/auth/request-password-reset",
		"/api/v1/user",
		"/api/v1/user/",
		"/api/v1/user/organizations",
		"/api/v1/user/name",
	}

	for _, authPath := range authPaths {
		if path == authPath || strings.HasPrefix(path, authPath+"/") {
			return true
		}
	}
	return false
}
