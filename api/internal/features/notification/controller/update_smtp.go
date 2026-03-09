package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
)

func (c *NotificationController) UpdateSmtp(f fuego.ContextWithBody[notification.UpdateSMTPConfigRequest]) (*types.MessageResponse, error) {
	SMTPConfigs, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	w, r := f.Response(), f.Request()

	jsonData, err := json.Marshal(SMTPConfigs)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	r.Body = io.NopCloser(bytes.NewBuffer(jsonData))

	if !c.parseAndValidate(w, r, &SMTPConfigs) {
		return nil, fuego.BadRequestError{
			Detail: "validation failed",
		}
	}

	err = c.service.UpdateSmtp(SMTPConfigs)
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
		Message: "SMTP configs updated successfully",
	}, nil
}
