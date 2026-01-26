package middleware

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"

	api_key_service "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	api_key_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// APIKeyAuthMiddleware is a middleware that authenticates requests using API keys.
// It extracts the API key from the Authorization header, verifies it, and adds
// the user and organization to the request context.
func APIKeyAuthMiddleware(next http.Handler, app *storage.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("ERROR APIKeyAuthMiddleware: Missing Authorization header for path %s", r.URL.Path)
			utils.SendErrorResponse(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix if present
		apiKeyString := strings.TrimPrefix(authHeader, "Bearer ")
		apiKeyString = strings.TrimSpace(apiKeyString)

		if apiKeyString == "" {
			log.Printf("ERROR APIKeyAuthMiddleware: Missing API key for path %s", r.URL.Path)
			utils.SendErrorResponse(w, "Unauthorized: missing API key", http.StatusUnauthorized)
			return
		}

		// Verify API key
		apiKeyStorage := api_key_storage.APIKeyStorage{
			DB:  app.Store.DB,
			Ctx: ctx,
		}
		apiKeyService := api_key_service.NewAPIKeyService(apiKeyStorage, logger.NewLogger())
		apiKey, err := apiKeyService.VerifyAPIKey(apiKeyString)
		if err != nil {
			log.Printf("ERROR APIKeyAuthMiddleware: Invalid API key for path %s: %v", r.URL.Path, err)
			utils.SendErrorResponse(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
			return
		}

		// Check if API key is valid (not revoked or expired)
		if !apiKey.IsValid() {
			log.Printf("ERROR APIKeyAuthMiddleware: API key is revoked or expired for path %s", r.URL.Path)
			utils.SendErrorResponse(w, "Unauthorized: API key is revoked or expired", http.StatusUnauthorized)
			return
		}

		// Get user from API key's user_id
		var user types.User
		err = app.Store.DB.NewSelect().
			Model(&user).
			Where("id = ?", apiKey.UserID).
			Scan(ctx)

		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("ERROR APIKeyAuthMiddleware: User not found for API key user_id %s", apiKey.UserID.String())
				utils.SendErrorResponse(w, "User not found", http.StatusUnauthorized)
				return
			}
			log.Printf("ERROR APIKeyAuthMiddleware: Failed to query user table: %v", err)
			utils.SendErrorResponse(w, "Failed to fetch user", http.StatusInternalServerError)
			return
		}

		// Compute backward compatibility fields
		user.ComputeCompatibilityFields()

		// Add user and organization to context
		ctx = context.WithValue(ctx, types.UserContextKey, &user)
		ctx = context.WithValue(ctx, types.OrganizationIDKey, apiKey.OrganizationID.String())

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
