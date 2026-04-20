package routes

import (
	"github.com/go-fuego/fuego"
	extension "github.com/nixopus/nixopus/api/internal/features/extension/controller"
)

func (router *Router) RegisterExtensionRoutes(extensionGroup *fuego.Server, extensionController *extension.ExtensionsController) {
	fuego.Get(
		extensionGroup,
		"",
		extensionController.GetExtensions,
		fuego.OptionSummary("List extensions"),
		fuego.OptionQuery("category", "Extension category filter"),
		fuego.OptionQuery("search", "Search term"),
		fuego.OptionQuery("type", "Extension type filter"),
		fuego.OptionQuery("sort_by", "Sort field"),
		fuego.OptionQuery("sort_dir", "Sort direction"),
		fuego.OptionQueryInt("page", "Page number"),
		fuego.OptionQueryInt("page_size", "Page size"),
	)
	fuego.Get(
		extensionGroup,
		"/categories",
		extensionController.GetCategories,
		fuego.OptionSummary("List extension categories"),
	)
	fuego.Get(
		extensionGroup,
		"/{id}",
		extensionController.GetExtension,
		fuego.OptionSummary("Get extension by ID"),
	)
	fuego.Get(
		extensionGroup,
		"/by-extension-id/{extension_id}",
		extensionController.GetExtensionByExtensionID,
		fuego.OptionSummary("Get extension by extension ID"),
	)
}
