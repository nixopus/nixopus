package routes

import (
	"github.com/go-fuego/fuego"
	auth "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
)

// RegisterAuthRoutes registers public authentication routes
// Most auth routes are handled by Better Auth
func (router *Router) RegisterAuthRoutes(authGroup *fuego.Server, authController *auth.AuthController) {
	// Check if admin is registered (public endpoint for registration flow)
	fuego.Get(authGroup, "/is-admin-registered", authController.IsAdminRegistered)
}

// RegisterAuthProtectedRoutes registers authenticated authentication routes
func (router *Router) RegisterAuthProtectedRoutes(authGroup *fuego.Server, authController *auth.AuthController) {
	// CLI init endpoint - requires authentication and creates a draft project
	fuego.Post(authGroup, "/cli/init", authController.HandleCLIInit)
}
