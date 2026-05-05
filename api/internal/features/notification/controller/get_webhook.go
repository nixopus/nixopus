package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *NotificationController) GetWebhookConfig(f fuego.ContextNoBody) (*types.WebhookConfigResponse, error) {
	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logNotificationDebug("GetWebhookConfig", "organization ID required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	webhookType := f.PathParam("type")

	ctxStr := fmt.Sprintf("org_id=%s webhook_type=%s", orgID, webhookType)
	c.logger.Log(logger.Info, "notification: GetWebhookConfig", ctxStr)

	config, err := c.service.GetWebhookConfig(f, &notification.GetWebhookConfigRequest{Type: webhookType}, orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("notification: GetWebhookConfig: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.WebhookConfigResponse{
		Status:  "success",
		Message: "Webhook config retrieved successfully",
		Data:    config,
	}, nil
}
