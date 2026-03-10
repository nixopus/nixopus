package routes

import (
	"github.com/go-fuego/fuego"
	domain "github.com/raghavyuva/nixopus-api/internal/features/domain/controller"
)

func (router *Router) RegisterDomainRoutes(domainGroup *fuego.Server, domainController *domain.DomainsController) {
	fuego.Get(
		domainGroup,
		"",
		domainController.GetDomains,
		fuego.OptionSummary("List domains"),
		fuego.OptionQuery("type", "Filter domains by type"),
	)
	fuego.Get(
		domainGroup,
		"/generate",
		domainController.GenerateRandomSubDomain,
		fuego.OptionSummary("Generate random subdomain"),
	)

	fuego.Post(
		domainGroup,
		"/custom",
		domainController.HandleAddCustomDomain,
		fuego.OptionSummary("Add custom domain"),
	)
	fuego.Delete(
		domainGroup,
		"/custom",
		domainController.HandleRemoveCustomDomain,
		fuego.OptionSummary("Remove custom domain"),
	)
	fuego.Post(
		domainGroup,
		"/verify",
		domainController.HandleVerifyCustomDomain,
		fuego.OptionSummary("Verify custom domain"),
	)
	fuego.Get(
		domainGroup,
		"/dns-check",
		domainController.HandleCheckDNSStatus,
		fuego.OptionSummary("Check custom domain DNS"),
		fuego.OptionQuery("id", "Custom domain ID", fuego.ParamRequired()),
	)
}
