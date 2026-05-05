package controller

import (
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *NotificationController) SendNotification(f fuego.ContextWithBody[notification.SendNotificationRequest]) (*types.SendNotificationResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logNotificationDebug("SendNotification", "body decode", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logNotificationDebug("SendNotification", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logNotificationDebug("SendNotification", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := fmt.Sprintf("org_id=%s user_id=%s channel=%s", orgID, user.ID, req.Channel)
	c.logger.Log(logger.Info, "notification: SendNotification", ctxStr)

	result := c.dispatcher.SendDirect(req, user.ID.String(), orgID.String())

	if !result.Success {
		c.logger.Log(logger.Warning, fmt.Sprintf("notification: SendNotification failed: %s", result.Error), ctxStr)
		return &types.SendNotificationResponse{
			Status:  "error",
			Message: result.Error,
			Data:    &result,
		}, nil
	}

	return &types.SendNotificationResponse{
		Status:  "success",
		Message: "Notification sent successfully",
		Data:    &result,
	}, nil
}
