package routes

import (
	"github.com/go-fuego/fuego"
	billingController "github.com/raghavyuva/nixopus-api/internal/features/billing/controller"
)

// RegisterBillingRoutes registers billing-related routes
func (router *Router) RegisterBillingRoutes(
	billingGroup *fuego.Server,
	controller *billingController.BillingController,
) {
	fuego.Get(billingGroup, "/status", controller.GetBillingStatus)
	fuego.Get(billingGroup, "/can-deploy", controller.CanDeploy)
	fuego.Post(billingGroup, "/checkout", controller.CreateCheckoutSession)
	fuego.Get(billingGroup, "/portal", controller.CreateBillingPortal)
	fuego.Get(billingGroup, "/invoices", controller.GetInvoices)
}
