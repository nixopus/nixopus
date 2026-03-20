package controller

import (
	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/controller/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *NotificationController) SendNotification(f fuego.ContextWithBody[notification.SendNotificationRequest]) (*types.SendNotificationResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	result := c.dispatcher.SendDirect(req, user.ID.String(), orgID.String())

	if !result.Success {
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
