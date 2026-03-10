package routes

import (
	"github.com/go-fuego/fuego"
	user "github.com/raghavyuva/nixopus-api/internal/features/user/controller"
)

// RegisterUserRoutes registers user-related routes
func (router *Router) RegisterUserRoutes(userGroup *fuego.Server, userController *user.UserController) {
	fuego.Get(userGroup, "", userController.GetUserDetails, fuego.OptionSummary("Get current user profile"))
	fuego.Patch(userGroup, "/name", userController.UpdateUserName, fuego.OptionSummary("Update user name"))
	fuego.Get(userGroup, "/settings", userController.GetSettings, fuego.OptionSummary("Get user settings"))
	fuego.Patch(userGroup, "/settings/font", userController.UpdateFont, fuego.OptionSummary("Update font settings"))
	fuego.Patch(userGroup, "/settings/theme", userController.UpdateTheme, fuego.OptionSummary("Update theme settings"))
	fuego.Patch(userGroup, "/settings/language", userController.UpdateLanguage, fuego.OptionSummary("Update language settings"))
	fuego.Patch(userGroup, "/settings/auto-update", userController.UpdateAutoUpdate, fuego.OptionSummary("Update auto-update settings"))
	fuego.Patch(userGroup, "/avatar", userController.UpdateAvatar, fuego.OptionSummary("Update user avatar"))
	fuego.Get(userGroup, "/preferences", userController.GetUserPreferences, fuego.OptionSummary("Get user preferences"))
	fuego.Put(userGroup, "/preferences", userController.UpdateUserPreferences, fuego.OptionSummary("Update user preferences"))
	fuego.Get(userGroup, "/onboarded", userController.GetIsOnboarded, fuego.OptionSummary("Check onboarding status"))
	fuego.Post(userGroup, "/onboarded", userController.MarkOnboardingComplete, fuego.OptionSummary("Mark onboarding complete"))
}
