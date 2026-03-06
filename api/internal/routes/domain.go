package routes

import (
	"github.com/go-fuego/fuego"
	domain "github.com/raghavyuva/nixopus-api/internal/features/domain/controller"
)

func (router *Router) RegisterDomainRoutes(domainGroup *fuego.Server, domainController *domain.DomainsController) {
	fuego.Get(domainGroup, "", domainController.GetDomains)
	fuego.Get(domainGroup, "/generate", domainController.GenerateRandomSubDomain)

	fuego.Post(domainGroup, "/custom", domainController.HandleAddCustomDomain)
	fuego.Delete(domainGroup, "/custom", domainController.HandleRemoveCustomDomain)
	fuego.Post(domainGroup, "/verify", domainController.HandleVerifyCustomDomain)
	fuego.Get(domainGroup, "/dns-check", domainController.HandleCheckDNSStatus)
}
