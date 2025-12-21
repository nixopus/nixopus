package routes

import (
	"github.com/go-fuego/fuego"
	domain "github.com/raghavyuva/nixopus-api/internal/features/domain/controller"
)

// RegisterDomainRoutes registers domain-related routes
func (router *Router) RegisterDomainRoutes(domainGroup *fuego.Server, domainsGroup *fuego.Server, domainController *domain.DomainsController) {
	fuego.Post(domainGroup, "", domainController.CreateDomain)
	fuego.Put(domainGroup, "", domainController.UpdateDomain)
	fuego.Delete(domainGroup, "", domainController.DeleteDomain)
	fuego.Get(domainGroup, "/generate", domainController.GenerateRandomSubDomain)
	fuego.Get(domainsGroup, "", domainController.GetDomains)
}
