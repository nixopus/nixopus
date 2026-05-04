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

func (c *NotificationController) DeleteWebhookConfig(f fuego.ContextWithBody[notification.DeleteWebhookConfigRequest]) (*types.MessageResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logNotificationDebug("DeleteWebhookConfig", "body decode", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logNotificationDebug("DeleteWebhookConfig", "organization ID required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := fmt.Sprintf("org_id=%s webhook_type=%s", orgID, req.Type)
	c.logger.Log(logger.Info, "notification: DeleteWebhookConfig", ctxStr)

	err = c.service.DeleteWebhookConfig(f, &req, orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("notification: DeleteWebhookConfig: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "Webhook config deleted successfully",
	}, nil
}
