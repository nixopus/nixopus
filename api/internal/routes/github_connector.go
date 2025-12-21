package routes

import (
	"github.com/go-fuego/fuego"
	githubConnector "github.com/raghavyuva/nixopus-api/internal/features/github-connector/controller"
)

// RegisterGithubConnectorRoutes registers GitHub connector routes
func (router *Router) RegisterGithubConnectorRoutes(githubGroup *fuego.Server, githubConnectorController *githubConnector.GithubConnectorController) {
	fuego.Post(githubGroup, "", githubConnectorController.CreateGithubConnector)
	fuego.Put(githubGroup, "", githubConnectorController.UpdateGithubConnectorRequest)
	fuego.Delete(githubGroup, "", githubConnectorController.DeleteGithubConnector)
	fuego.Get(githubGroup, "/all", githubConnectorController.GetGithubConnectors)
	fuego.Get(githubGroup, "/repositories", githubConnectorController.GetGithubRepositories)
	fuego.Post(githubGroup, "/repository/branches", githubConnectorController.GetGithubRepositoryBranches)
}
