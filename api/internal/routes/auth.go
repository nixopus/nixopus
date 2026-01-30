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
