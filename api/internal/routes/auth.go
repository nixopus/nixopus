package routes

import (
	"github.com/go-fuego/fuego"
	auth "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
)

// RegisterAuthRoutes registers public authentication routes
func (router *Router) RegisterAuthRoutes(authGroup *fuego.Server, authController *auth.AuthController) {
	fuego.Get(authGroup, "/is-admin-registered", authController.IsAdminRegistered)
}

// RegisterAuthenticatedAuthRoutes registers protected authentication routes
func (router *Router) RegisterAuthenticatedAuthRoutes(authGroup *fuego.Server, authController *auth.AuthController) {
	fuego.Post(authGroup, "/logout", authController.Logout)
	fuego.Post(authGroup, "/send-verification-email", authController.SendVerificationEmail)
	fuego.Get(authGroup, "/verify-email", authController.VerifyEmail)
	fuego.Post(authGroup, "/create-user", authController.CreateUser)
	fuego.Post(authGroup, "/setup-2fa", authController.SetupTwoFactor)
	fuego.Post(authGroup, "/verify-2fa", authController.VerifyTwoFactor)
	fuego.Post(authGroup, "/disable-2fa", authController.DisableTwoFactor)
	fuego.Post(authGroup, "/2fa-login", authController.TwoFactorLogin)
	fuego.Post(authGroup, "/api-keys", authController.CreateAPIKey)
	fuego.Get(authGroup, "/api-keys", authController.ListAPIKeys)
	fuego.Delete(authGroup, "/api-keys/{id}", authController.RevokeAPIKey)
}
