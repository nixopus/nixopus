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

func (c *NotificationController) CreateWebhookConfig(f fuego.ContextWithBody[notification.CreateWebhookConfigRequest]) (*types.WebhookConfigResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logNotificationDebug("CreateWebhookConfig", "body decode", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logNotificationDebug("CreateWebhookConfig", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logNotificationDebug("CreateWebhookConfig", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := fmt.Sprintf("org_id=%s user_id=%s webhook_type=%s", orgID, user.ID, req.Type)
	c.logger.Log(logger.Info, "notification: CreateWebhookConfig", ctxStr)

	config, err := c.service.CreateWebhookConfig(f, &req, user.ID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("notification: CreateWebhookConfig: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.WebhookConfigResponse{
		Status:  "success",
		Message: "Webhook config created successfully",
		Data:    config,
	}, nil
}
