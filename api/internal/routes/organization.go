package routes

import (
	"github.com/go-fuego/fuego"
	organization "github.com/raghavyuva/nixopus-api/internal/features/organization/controller"
)

// RegisterOrganizationRoutes registers organization routes
func (router *Router) RegisterOrganizationRoutes(organizationGroup *fuego.Server, organizationController *organization.OrganizationsController) {
	fuego.Get(organizationGroup, "/users", organizationController.GetOrganizationUsers)
	fuego.Post(organizationGroup, "/remove-user", organizationController.RemoveUserFromOrganization)
	fuego.Post(organizationGroup, "/update-user-role", organizationController.UpdateUserRole)
	fuego.Put(organizationGroup, "", organizationController.UpdateOrganization)
	fuego.Post(organizationGroup, "", organizationController.CreateOrganization)
	fuego.Delete(organizationGroup, "", organizationController.DeleteOrganization)
	fuego.Get(organizationGroup, "", organizationController.GetOrganization)
	fuego.Get(organizationGroup, "/all", organizationController.GetOrganizations)
	fuego.Post(organizationGroup, "/invite/send", organizationController.SendInvite)
	fuego.Post(organizationGroup, "/invite/resend", organizationController.ResendInvite)
}
