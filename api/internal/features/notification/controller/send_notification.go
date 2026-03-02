package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/controller/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *NotificationController) SendNotification(f fuego.ContextWithBody[notification.SendNotificationRequest]) (*types.SendNotificationResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	result := c.notification.SendDirectNotification(req, user.ID.String(), orgID.String())

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
