package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
)

func (c *NotificationController) DeleteSmtp(f fuego.ContextWithBody[notification.DeleteSMTPConfigRequest]) (*types.MessageResponse, error) {
	w, r := f.Response(), f.Request()

	var SMTPConfigs notification.DeleteSMTPConfigRequest
	if !c.parseAndValidate(w, r, &SMTPConfigs) {
		return nil, fuego.BadRequestError{
			Detail: "validation failed",
		}
	}

	err := c.service.DeleteSmtp(SMTPConfigs.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "SMTP configs deleted successfully",
	}, nil
}
