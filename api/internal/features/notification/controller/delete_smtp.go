package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/controller/types"
)

func (c *NotificationController) DeleteSmtp(f fuego.ContextWithBody[notification.DeleteSMTPConfigRequest]) (*types.MessageResponse, error) {
	SMTPConfigs, err := f.Body()

	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()

	if !c.parseAndValidate(w, r, &SMTPConfigs) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	err = c.service.DeleteSmtp(SMTPConfigs.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "SMTP configs deleted successfully",
	}, nil
}
