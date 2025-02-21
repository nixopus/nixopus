package auth

import "github.com/raghavyuva/nixopus-api/internal/storage"

type AuthController struct {
	app *storage.App
}

// NewAuthController creates a new AuthController with the given App.
//
// This function creates a new AuthController with the given App and returns a pointer to it.
//
// The App passed to this function should be a valid App that has been created with storage.NewApp.
func NewAuthController(app *storage.App) *AuthController {
	return &AuthController{
		app: app,
	}
}
