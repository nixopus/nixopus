package routes

import (
	"github.com/go-fuego/fuego"
	auth "github.com/nixopus/nixopus/api/internal/features/auth/controller"
)

// RegisterAuthRoutes registers public authentication routes
// Most auth routes are handled by Better Auth
func (router *Router) RegisterAuthRoutes(authGroup *fuego.Server, authController *auth.AuthController) {
	// Check if admin is registered (public endpoint for registration flow)
	fuego.Get(
		authGroup,
		"/is-admin-registered",
		authController.IsAdminRegistered,
		fuego.OptionSummary("Check admin registration"),
	)
}

// RegisterAuthProtectedRoutes registers authenticated authentication routes
func (router *Router) RegisterAuthProtectedRoutes(authGroup *fuego.Server, authController *auth.AuthController) {
	// Bootstrap: user, orgs, activeOrgId, isOnboarded, provisionStatus, hasServers (skips org resolution)
	fuego.Get(
		authGroup,
		"/bootstrap",
		authController.HandleBootstrap,
		fuego.OptionSummary("Get bootstrap session data"),
	)
}
