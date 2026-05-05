package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
	"github.com/nixopus/nixopus/api/internal/utils"
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

	user := utils.GetUser(w, r)
	ctxStr := notificationRequestLogData(r, user)
	if SMTPConfigs.ID != uuid.Nil {
		if ctxStr != "" {
			ctxStr = fmt.Sprintf("%s smtp_config_id=%s", ctxStr, SMTPConfigs.ID)
		} else {
			ctxStr = fmt.Sprintf("smtp_config_id=%s", SMTPConfigs.ID)
		}
	}
	c.logger.Log(logger.Info, "notification: UpdateSmtp", ctxStr)

	err = c.service.UpdateSmtp(SMTPConfigs)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("notification: UpdateSmtp: %v", err), ctxStr)
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
