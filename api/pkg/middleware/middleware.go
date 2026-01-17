// Package middleware provides public access to middleware functions from the internal package.
// This allows other modules (like cloud) to use the middleware without importing internal packages.
package middleware

import (
	"net/http"

	internalLogger "github.com/raghavyuva/nixopus-api/internal/features/logger"
	internalMiddleware "github.com/raghavyuva/nixopus-api/internal/middleware"
	"github.com/raghavyuva/nixopus-api/pkg/cache"
	"github.com/raghavyuva/nixopus-api/pkg/storage"
)

// Logger is a type alias for the internal logger.Logger type.
// This allows other modules to reference the Logger type without importing internal packages.
type Logger = internalLogger.Logger

// NewLogger creates a new Logger instance.
func NewLogger() Logger {
	return internalLogger.NewLogger()
}

// RecoveryMiddleware recovers from panics and returns 500 errors.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return internalMiddleware.RecoveryMiddleware(next)
}

// LoggingMiddleware logs HTTP requests with colored output.
func LoggingMiddleware(next http.Handler) http.Handler {
	return internalMiddleware.LoggingMiddleware(next)
}

// CorsMiddleware sets the necessary CORS headers for the response.
func CorsMiddleware(next http.Handler) http.Handler {
	return internalMiddleware.CorsMiddleware(next)
}

// SupertokensCorsMiddleware handles SuperTokens CORS requirements.
func SupertokensCorsMiddleware(next http.Handler) http.Handler {
	return internalMiddleware.SupertokensCorsMiddleware(next)
}

// AuthMiddleware checks if the request has a valid SuperTokens session.
// It requires storage.App and cache.Cache instances.
func AuthMiddleware(next http.Handler, app *storage.App, cacheInstance *cache.Cache) http.Handler {
	return internalMiddleware.AuthMiddleware(next, app, cacheInstance)
}

// RBACMiddleware validates SuperTokens permission claims for the given resource.
// It requires storage.App and a resource name.
func RBACMiddleware(next http.Handler, app *storage.App, resourceName string) http.Handler {
	return internalMiddleware.RBACMiddleware(next, app, resourceName)
}

// InitRBACCache initializes the RBAC cache with a cache instance.
func InitRBACCache(c *cache.Cache) {
	internalMiddleware.InitRBACCache(c)
}

// AuditMiddleware captures audit logs for all authenticated requests.
// It requires storage.App, logger, and resource type.
func AuditMiddleware(next http.Handler, app *storage.App, l Logger, resourceType string) http.Handler {
	return internalMiddleware.AuditMiddleware(next, app, l, resourceType)
}
