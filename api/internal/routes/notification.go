package routes

import (
	"github.com/go-fuego/fuego"
	notificationController "github.com/raghavyuva/nixopus-api/internal/features/notification/controller"
)

// RegisterNotificationRoutes registers notification routes
func (router *Router) RegisterNotificationRoutes(notificationGroup *fuego.Server, notificationController *notificationController.NotificationController) {
	smtpGroup := fuego.Group(notificationGroup, "/smtp")
	fuego.Post(smtpGroup, "", notificationController.AddSmtp)
	fuego.Get(smtpGroup, "", notificationController.GetSmtp)
	fuego.Put(smtpGroup, "", notificationController.UpdateSmtp)
	fuego.Delete(smtpGroup, "", notificationController.DeleteSmtp)

	preferenceGroup := fuego.Group(notificationGroup, "/preferences")
	fuego.Post(preferenceGroup, "", notificationController.UpdatePreference)
	fuego.Get(preferenceGroup, "", notificationController.GetPreferences)

	webhookGroup := fuego.Group(notificationGroup, "/webhook")
	fuego.Post(webhookGroup, "", notificationController.CreateWebhookConfig)
	fuego.Get(webhookGroup, "/{type}", notificationController.GetWebhookConfig)
	fuego.Put(webhookGroup, "", notificationController.UpdateWebhookConfig)
	fuego.Delete(webhookGroup, "", notificationController.DeleteWebhookConfig)
}
