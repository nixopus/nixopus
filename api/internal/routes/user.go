package routes

import (
	"github.com/go-fuego/fuego"
	user "github.com/raghavyuva/nixopus-api/internal/features/user/controller"
)

// RegisterUserRoutes registers user-related routes
func (router *Router) RegisterUserRoutes(userGroup *fuego.Server, userController *user.UserController) {
	fuego.Get(userGroup, "", userController.GetUserDetails)
	fuego.Patch(userGroup, "/name", userController.UpdateUserName)
	fuego.Get(userGroup, "/organizations", userController.GetUserOrganizations)
	fuego.Get(userGroup, "/settings", userController.GetSettings)
	fuego.Patch(userGroup, "/settings/font", userController.UpdateFont)
	fuego.Patch(userGroup, "/settings/theme", userController.UpdateTheme)
	fuego.Patch(userGroup, "/settings/language", userController.UpdateLanguage)
	fuego.Patch(userGroup, "/settings/auto-update", userController.UpdateAutoUpdate)
	fuego.Patch(userGroup, "/avatar", userController.UpdateAvatar)
	fuego.Get(userGroup, "/preferences", userController.GetUserPreferences)
	fuego.Put(userGroup, "/preferences", userController.UpdateUserPreferences)
}
