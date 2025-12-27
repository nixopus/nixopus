package routes

import (
	"github.com/go-fuego/fuego"
	extension "github.com/raghavyuva/nixopus-api/internal/features/extension/controller"
)

// RegisterExtensionRoutes registers extension routes
func (router *Router) RegisterExtensionRoutes(extensionGroup *fuego.Server, extensionController *extension.ExtensionsController) {
	fuego.Get(extensionGroup, "", extensionController.GetExtensions)
	fuego.Get(extensionGroup, "/categories", extensionController.GetCategories)
	fuego.Get(extensionGroup, "/{id}", extensionController.GetExtension)
	fuego.Get(extensionGroup, "/by-extension-id/{extension_id}", extensionController.GetExtensionByExtensionID)
	fuego.Get(extensionGroup, "/by-extension-id/{extension_id}/executions", extensionController.ListExecutionsByExtensionID)
	fuego.Post(extensionGroup, "/{extension_id}/run", extensionController.RunExtension)
	fuego.Post(extensionGroup, "/execution/{execution_id}/cancel", extensionController.CancelExecution)
	fuego.Get(extensionGroup, "/execution/{execution_id}", extensionController.GetExecution)
	fuego.Get(extensionGroup, "/execution/{execution_id}/logs", extensionController.ListExecutionLogs)
	fuego.Post(extensionGroup, "/{extension_id}/fork", extensionController.ForkExtension)
	fuego.Delete(extensionGroup, "/{id}", extensionController.DeleteFork)
}
