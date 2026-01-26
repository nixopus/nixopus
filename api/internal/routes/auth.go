package routes

import (
	"net/http"

	"github.com/go-fuego/fuego"
	auth "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
	"github.com/raghavyuva/nixopus-api/internal/middleware"
)

// RegisterAuthRoutes registers public authentication routes
// Most auth routes are handled by Better Auth - only keeping API key routes
func (router *Router) RegisterAuthRoutes(authGroup *fuego.Server, authController *auth.AuthController) {
	// Apply rate limiting middleware to the auth group to prevent abuse
	fuego.Use(authGroup, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Apply rate limiting only to the validation endpoint
			if r.URL.Path == "/api/v1/auth/validate-api-key" && r.Method == http.MethodPost {
				middleware.RateLimiter(next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Public API key validation endpoint (rate limited for security)
	fuego.Post(authGroup, "/validate-api-key", authController.ValidateAPIKey)

	// CLI init endpoint - creates a draft project using API key authentication
	fuego.Post(authGroup, "/cli-init", authController.HandleCLIInit)

	// Check if admin is registered (public endpoint for registration flow)
	fuego.Get(authGroup, "/is-admin-registered", authController.IsAdminRegistered)
}

// RegisterAuthenticatedAuthRoutes registers protected authentication routes
// Only API key management is kept - Better Auth handles login/logout/2FA/etc
func (router *Router) RegisterAuthenticatedAuthRoutes(authGroup *fuego.Server, authController *auth.AuthController) {
	fuego.Post(authGroup, "/api-keys", authController.CreateAPIKey)
	fuego.Get(authGroup, "/api-keys", authController.ListAPIKeys)
	fuego.Delete(authGroup, "/api-keys/{id}", authController.RevokeAPIKey)
}
