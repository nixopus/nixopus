package routes

import (
	"github.com/go-fuego/fuego"
	extension "github.com/raghavyuva/nixopus-api/internal/features/extension/controller"
)

// RegisterExtensionRoutes registers extension routes
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
	fuego.Get(
		extensionGroup,
		"/by-extension-id/{extension_id}/executions",
		extensionController.ListExecutionsByExtensionID,
		fuego.OptionSummary("List extension executions"),
	)
	fuego.Post(
		extensionGroup,
		"/{extension_id}/run",
		extensionController.RunExtension,
		fuego.OptionSummary("Run extension"),
	)
	fuego.Post(
		extensionGroup,
		"/execution/{execution_id}/cancel",
		extensionController.CancelExecution,
		fuego.OptionSummary("Cancel execution"),
	)
	fuego.Get(
		extensionGroup,
		"/execution/{execution_id}",
		extensionController.GetExecution,
		fuego.OptionSummary("Get execution"),
	)
	fuego.Get(
		extensionGroup,
		"/execution/{execution_id}/logs",
		extensionController.ListExecutionLogs,
		fuego.OptionSummary("List execution logs"),
		fuego.OptionQueryInt("afterSeq", "Return logs after this sequence"),
		fuego.OptionQueryInt("limit", "Maximum logs to return"),
	)
	fuego.Post(
		extensionGroup,
		"/{extension_id}/fork",
		extensionController.ForkExtension,
		fuego.OptionSummary("Fork extension"),
	)
	fuego.Delete(
		extensionGroup,
		"/{id}",
		extensionController.DeleteFork,
		fuego.OptionSummary("Delete forked extension"),
	)
}
