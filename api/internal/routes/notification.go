package routes

import (
	"github.com/go-fuego/fuego"
	notificationController "github.com/nixopus/nixopus/api/internal/features/notification/controller"
)

// RegisterNotificationRoutes registers notification routes
func (router *Router) RegisterNotificationRoutes(notificationGroup *fuego.Server, notificationController *notificationController.NotificationController) {
	smtpGroup := fuego.Group(notificationGroup, "/smtp")
	fuego.Post(
		smtpGroup,
		"",
		notificationController.AddSmtp,
		fuego.OptionSummary("Create SMTP config"),
	)
	fuego.Get(
		smtpGroup,
		"",
		notificationController.GetSmtp,
		fuego.OptionSummary("Get SMTP config"),
		fuego.OptionQuery("id", "Organization ID", fuego.ParamRequired()),
	)
	fuego.Put(
		smtpGroup,
		"",
		notificationController.UpdateSmtp,
		fuego.OptionSummary("Update SMTP config"),
	)
	fuego.Delete(
		smtpGroup,
		"",
		notificationController.DeleteSmtp,
		fuego.OptionSummary("Delete SMTP config"),
	)

	preferenceGroup := fuego.Group(notificationGroup, "/preferences")
	fuego.Post(
		preferenceGroup,
		"",
		notificationController.UpdatePreference,
		fuego.OptionSummary("Update notification preferences"),
	)
	fuego.Get(
		preferenceGroup,
		"",
		notificationController.GetPreferences,
		fuego.OptionSummary("Get notification preferences"),
	)

	webhookGroup := fuego.Group(notificationGroup, "/webhook")
	fuego.Post(
		webhookGroup,
		"",
		notificationController.CreateWebhookConfig,
		fuego.OptionSummary("Create webhook config"),
	)
	fuego.Get(
		webhookGroup,
		"/{type}",
		notificationController.GetWebhookConfig,
		fuego.OptionSummary("Get webhook config"),
	)
	fuego.Put(
		webhookGroup,
		"",
		notificationController.UpdateWebhookConfig,
		fuego.OptionSummary("Update webhook config"),
	)
	fuego.Delete(
		webhookGroup,
		"",
		notificationController.DeleteWebhookConfig,
		fuego.OptionSummary("Delete webhook config"),
	)

	fuego.Post(
		notificationGroup,
		"/send",
		notificationController.SendNotification,
		fuego.OptionSummary("Send notification"),
	)
}
