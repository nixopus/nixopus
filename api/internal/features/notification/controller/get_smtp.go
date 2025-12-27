package controller

import (
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/controller/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *NotificationController) GetSmtp(f fuego.ContextNoBody) (*types.SMTPConfigResponse, error) {
	id := f.QueryParam("id")
	if id == "" {
		return nil, fuego.HTTPError{
			Err:    notification.ErrMissingID,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    notification.ErrAccessDenied,
			Status: http.StatusUnauthorized,
		}
	}

	orgID := ""
	for _, org := range user.Organizations {
		if org.ID.String() == id {
			orgID = org.ID.String()
			break
		}
	}

	if orgID == "" {
		return nil, fuego.HTTPError{
			Err:    notification.ErrUserDoesNotBelongToOrganization,
			Status: http.StatusForbidden,
		}
	}

	SMTPConfigs, err := c.service.GetSmtp(user.ID.String(), orgID)
	if err != nil {
		if errors.Is(err, notification.ErrSMTPConfigNotFound) {
			return smtpNotFoundResp(), nil
		}

		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	if SMTPConfigs == nil {
		return smtpNotFoundResp(), nil
	}

	return &types.SMTPConfigResponse{
		Status:  "success",
		Message: "SMTP configs fetched successfully",
		Data:    SMTPConfigs,
	}, nil
}

func smtpNotFoundResp() *types.SMTPConfigResponse {
	return &types.SMTPConfigResponse{
		Status:  "success",
		Message: "No SMTP configs were found",
		Data:    nil,
	}
}
