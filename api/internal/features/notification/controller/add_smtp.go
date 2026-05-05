package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *NotificationController) AddSmtp(f fuego.ContextWithBody[notification.CreateSMTPConfigRequest]) (*types.MessageResponse, error) {
	w, r := f.Response(), f.Request()

	var SMTPConfigs notification.CreateSMTPConfigRequest
	if !c.parseAndValidate(w, r, &SMTPConfigs) {
		return nil, fuego.BadRequestError{
			Detail: "validation failed",
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		c.logNotificationDebug("AddSmtp", "authentication required", notificationRequestLogData(r, nil))
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := fmt.Sprintf("org_id=%s user_id=%s", SMTPConfigs.OrganizationID, user.ID)
	c.logger.Log(logger.Info, "notification: AddSmtp", ctxStr)

	err := c.service.AddSmtp(SMTPConfigs, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("notification: AddSmtp: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "SMTP Config added successfully",
	}, nil
}
