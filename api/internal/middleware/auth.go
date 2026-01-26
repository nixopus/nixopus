package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/cache"
	api_key_service "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	api_key_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	betterauth "github.com/raghavyuva/nixopus-api/internal/features/betterauth"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"github.com/uptrace/bun"
)

// Member represents the member table from Better Auth schema (auth_schema.ts)
// Table: "member" with columns: id, organization_id, user_id, role, created_at
type Member struct {
	bun.BaseModel  `bun:"table:member,alias:m"`
	ID             uuid.UUID `bun:"id,pk,type:uuid"`
	OrganizationID uuid.UUID `bun:"organization_id,type:uuid,notnull"`
	UserID         uuid.UUID `bun:"user_id,type:uuid,notnull"`
	Role           string    `bun:"role,type:text,notnull,default:'member'"`
	CreatedAt      time.Time `bun:"created_at,type:timestamp,notnull"`
}

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
				// Get organization ID from Better Auth session or header
				organizationID, err := utils.GetOrganizationIDString(ctx, r, app)
				if err != nil || organizationID == "" {
					// If no organization ID provided, use the first organization for the user
					firstOrgID, err := getFirstUserOrganization(ctx, app, userIDUUID)
					if err != nil || firstOrgID == "" {
						log.Printf("WARN AuthMiddleware: No organization ID provided and user has no organizations")
						utils.SendErrorResponse(w, "No organization ID provided and user has no organizations", http.StatusBadRequest)
						return
					}
					organizationID = firstOrgID
					log.Printf("INFO AuthMiddleware: Using first organization %s for user %s", organizationID, betterAuthUserID)
				}

				// Check organization membership via Better Auth API
				// Use Better Auth user ID for membership check
				var belongsToOrg bool
				if cache != nil {
					if cached, err := cache.GetOrgMembership(ctx, betterAuthUserID, organizationID); err == nil {
						belongsToOrg = cached
					}
				}

				if !belongsToOrg {
					// Verify membership via Better Auth API
					member, err := getBetterAuthOrganizationMember(ctx, r, betterAuthUserID, organizationID)
					if err != nil || member == nil {
						log.Printf("WARN AuthMiddleware: User %s (Better Auth ID: %s) does not belong to organization %s: %v", user.ID.String(), betterAuthUserID, organizationID, err)
						utils.SendErrorResponse(w, "User does not belong to the specified organization", http.StatusForbidden)
						return
					}
					belongsToOrg = true

					// Cache the membership
					if cache != nil {
						cache.SetOrgMembership(ctx, betterAuthUserID, organizationID, belongsToOrg)
					}
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

// getFirstUserOrganization gets the first organization ID for a user from the Better Auth member table
func getFirstUserOrganization(ctx context.Context, app *storage.App, userID uuid.UUID) (string, error) {
	var member Member
	err := app.Store.DB.NewSelect().
		Model(&member).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user has no organizations")
		}
		return "", fmt.Errorf("failed to query member table: %w", err)
	}

	return member.OrganizationID.String(), nil
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
