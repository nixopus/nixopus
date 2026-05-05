package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *NotificationController) DeleteSmtp(f fuego.ContextWithBody[notification.DeleteSMTPConfigRequest]) (*types.MessageResponse, error) {
	w, r := f.Response(), f.Request()

	var SMTPConfigs notification.DeleteSMTPConfigRequest
	if !c.parseAndValidate(w, r, &SMTPConfigs) {
		return nil, fuego.BadRequestError{
			Detail: "validation failed",
		}
	}

	ctxStr := notificationRequestLogData(r, utils.GetUser(w, r))
	if SMTPConfigs.ID != uuid.Nil {
		if ctxStr != "" {
			ctxStr = fmt.Sprintf("%s smtp_config_id=%s", ctxStr, SMTPConfigs.ID)
		} else {
			ctxStr = fmt.Sprintf("smtp_config_id=%s", SMTPConfigs.ID)
		}
	}
	c.logger.Log(logger.Info, "notification: DeleteSmtp", ctxStr)

	err := c.service.DeleteSmtp(SMTPConfigs.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("notification: DeleteSmtp: %v", err), ctxStr)
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: status,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "SMTP configs deleted successfully",
	}, nil
}
