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

func (c *NotificationController) UpdateWebhookConfig(f fuego.ContextWithBody[notification.UpdateWebhookConfigRequest]) (*types.WebhookConfigResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logNotificationDebug("UpdateWebhookConfig", "body decode", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logNotificationDebug("UpdateWebhookConfig", "organization ID required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := fmt.Sprintf("org_id=%s webhook_type=%s", orgID, req.Type)
	c.logger.Log(logger.Info, "notification: UpdateWebhookConfig", ctxStr)

	config, err := c.service.UpdateWebhookConfig(f, &req, orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("notification: UpdateWebhookConfig: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.WebhookConfigResponse{
		Status:  "success",
		Message: "Webhook config updated successfully",
		Data:    config,
	}, nil
}
