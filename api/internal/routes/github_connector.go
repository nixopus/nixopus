package routes

import (
	"github.com/go-fuego/fuego"
	githubConnector "github.com/raghavyuva/nixopus-api/internal/features/github-connector/controller"
)

// RegisterGithubConnectorRoutes registers GitHub connector routes
func (router *Router) RegisterGithubConnectorRoutes(githubGroup *fuego.Server, githubConnectorController *githubConnector.GithubConnectorController) {
	fuego.Post(githubGroup, "", githubConnectorController.CreateGithubConnector, fuego.OptionSummary("Create GitHub connector"))
	fuego.Put(githubGroup, "", githubConnectorController.UpdateGithubConnectorRequest, fuego.OptionSummary("Update GitHub connector"))
	fuego.Delete(githubGroup, "", githubConnectorController.DeleteGithubConnector, fuego.OptionSummary("Delete GitHub connector"))
	fuego.Get(githubGroup, "/all", githubConnectorController.GetGithubConnectors, fuego.OptionSummary("List GitHub connectors"))
	fuego.Get(githubGroup, "/repositories", githubConnectorController.GetGithubRepositories, fuego.OptionSummary("List GitHub repositories"))
	fuego.Post(githubGroup, "/repository/branches", githubConnectorController.GetGithubRepositoryBranches, fuego.OptionSummary("List repository branches"))
}
