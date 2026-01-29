package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/cache"
	api_key_service "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	api_key_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	betterauth "github.com/raghavyuva/nixopus-api/internal/features/betterauth"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// AuthMiddleware is a middleware that checks if the request has a valid
// Better Auth session. If the session is valid, it adds both the user and
// the authenticated client to the request context.
// If session auth fails, it falls back to API key authentication.
func AuthMiddleware(next http.Handler, app *storage.App, cache *cache.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Try Better Auth session first
		sessionResp, err := betterauth.VerifySession(r)
		var usingAPIKey bool
		if err != nil {
			// Session auth failed, try API key as fallback
			log.Printf("INFO AuthMiddleware: Session auth failed for path %s, trying API key: %v", r.URL.Path, err)
			if apiKeyCtx := tryAPIKeyAuth(r, app, ctx); apiKeyCtx != nil {
				// API key auth succeeded, use that context
				ctx = apiKeyCtx
				r = r.WithContext(ctx)
				usingAPIKey = true
				// Skip organization membership check for API keys (they're already scoped to org)
			} else {
				// Both auth methods failed
				log.Printf("ERROR AuthMiddleware: Both session and API key auth failed for path %s", r.URL.Path)
				utils.SendErrorResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}
		} else {
			// Session auth succeeded, continue with normal flow
			betterAuthUserID := sessionResp.User.ID

			// Optionally cache can be disabled by setting the X-Disable-Cache header to true
			disableCache := r.Header.Get("X-Disable-Cache")
			if disableCache == "true" {
				cache = nil
			}

			// Get user details from Better Auth user table (matching auth_schema.ts)
			var user types.User
			userIDUUID, err := uuid.Parse(betterAuthUserID)
			if err != nil {
				log.Printf("ERROR AuthMiddleware: Invalid Better Auth user ID format: %s", betterAuthUserID)
				utils.SendErrorResponse(w, "Invalid user ID", http.StatusUnauthorized)
				return
			}

			err = app.Store.DB.NewSelect().
				Model(&user).
				Where("id = ?", userIDUUID).
				Scan(ctx)

			if err != nil {
				if err == sql.ErrNoRows {
					log.Printf("ERROR AuthMiddleware: User not found in Better Auth user table with ID %s", betterAuthUserID)
					utils.SendErrorResponse(w, "User not found", http.StatusUnauthorized)
					return
				}
				log.Printf("ERROR AuthMiddleware: Failed to query Better Auth user table: %v", err)
				utils.SendErrorResponse(w, "Failed to fetch user", http.StatusInternalServerError)
				return
			}

			// Compute backward compatibility fields
			user.ComputeCompatibilityFields()

			// Use user directly in context
			ctx = context.WithValue(ctx, types.UserContextKey, &user)

			if !isAuthEndpoint(r.URL.Path) {
				// Resolve and verify organization membership
				organizationID, err := resolveAndVerifyOrganization(ctx, r, cache, betterAuthUserID)
				if err != nil {
					log.Printf("ERROR AuthMiddleware: Failed to resolve organization: %v", err)
					statusCode := http.StatusBadRequest
					if strings.Contains(err.Error(), "does not belong") {
						statusCode = http.StatusForbidden
					}
					utils.SendErrorResponse(w, err.Error(), statusCode)
					return
				}

				ctx = context.WithValue(ctx, types.OrganizationIDKey, organizationID)
			}
		}

		// For API key auth, organization is already set in context, skip membership check
		if usingAPIKey && !isAuthEndpoint(r.URL.Path) {
			// Organization ID is already set from API key, no need to check membership
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
		"/api/v1/auth/api-keys",
		"/api/v1/user",
		"/api/v1/user/",
		"/api/v1/user/organizations",
		"/api/v1/user/name",
		// Better Auth endpoints
		"/api/auth",
	}

	for _, authPath := range authPaths {
		if path == authPath || strings.HasPrefix(path, authPath+"/") {
			return true
		}
	}
	return false
}

// resolveAndVerifyOrganization resolves the organization ID from Better Auth session
// and verifies user membership via Better Auth API (with caching).
func resolveAndVerifyOrganization(
	ctx context.Context,
	r *http.Request,
	cache *cache.Cache,
	betterAuthUserID string,
) (string, error) {
	// Get organization ID from Better Auth session only
	organizationID, err := utils.GetOrganizationIDFromBetterAuth(r)
	if err != nil {
		return "", fmt.Errorf("failed to get organization ID from Better Auth session: %w", err)
	}

	if organizationID == "" {
		return "", fmt.Errorf("no organization ID found in Better Auth session")
	}

	// Check organization membership via Better Auth API (with caching)
	belongsToOrg, err := verifyOrganizationMembership(ctx, r, cache, betterAuthUserID, organizationID)
	if err != nil {
		return "", fmt.Errorf("user %s does not belong to organization %s: %w", betterAuthUserID, organizationID, err)
	}

	if !belongsToOrg {
		return "", fmt.Errorf("user %s does not belong to organization %s", betterAuthUserID, organizationID)
	}

	return organizationID, nil
}

// verifyOrganizationMembership verifies if a user belongs to an organization.
// Uses cache to avoid repeated API calls.
func verifyOrganizationMembership(
	ctx context.Context,
	r *http.Request,
	cache *cache.Cache,
	betterAuthUserID string,
	organizationID string,
) (bool, error) {
	// Check cache first
	if cache != nil {
		if cached, err := cache.GetOrgMembership(ctx, betterAuthUserID, organizationID); err == nil && cached {
			return true, nil
		}
	}

	// Verify membership via Better Auth API
	member, err := getBetterAuthOrganizationMember(ctx, r, betterAuthUserID, organizationID)
	if err != nil || member == nil {
		// Cache negative result to avoid repeated API calls
		if cache != nil {
			cache.SetOrgMembership(ctx, betterAuthUserID, organizationID, false)
		}
		return false, err
	}

	// Cache positive result
	if cache != nil {
		cache.SetOrgMembership(ctx, betterAuthUserID, organizationID, true)
	}

	return true, nil
}

// tryAPIKeyAuth attempts to authenticate using API key from Authorization header
// Returns context with user and organization if successful, nil otherwise
func tryAPIKeyAuth(r *http.Request, app *storage.App, ctx context.Context) context.Context {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	// Remove "Bearer " prefix if present
	apiKeyString := strings.TrimPrefix(authHeader, "Bearer ")
	apiKeyString = strings.TrimSpace(apiKeyString)

	if apiKeyString == "" {
		return nil
	}

	// Verify API key
	apiKeyStorage := api_key_storage.APIKeyStorage{
		DB:  app.Store.DB,
		Ctx: ctx,
	}
	apiKeyService := api_key_service.NewAPIKeyService(apiKeyStorage, logger.NewLogger())
	apiKey, err := apiKeyService.VerifyAPIKey(apiKeyString)
	if err != nil {
		return nil
	}

	// Check if API key is valid (not revoked or expired)
	if !apiKey.IsValid() {
		return nil
	}

	// Get user from API key's user_id
	var user types.User
	err = app.Store.DB.NewSelect().
		Model(&user).
		Where("id = ?", apiKey.UserID).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR AuthMiddleware: User not found for API key user_id %s", apiKey.UserID.String())
			return nil
		}
		log.Printf("ERROR AuthMiddleware: Failed to query user table for API key: %v", err)
		return nil
	}

	// Compute backward compatibility fields
	user.ComputeCompatibilityFields()

	// Add user and organization to context
	ctx = context.WithValue(ctx, types.UserContextKey, &user)
	ctx = context.WithValue(ctx, types.OrganizationIDKey, apiKey.OrganizationID.String())

	return ctx
}
