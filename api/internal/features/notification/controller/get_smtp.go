package controller

import (
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *NotificationController) GetSmtp(f fuego.ContextNoBody) (*types.SMTPConfigResponse, error) {
	id := f.QueryParam("id")
	if id == "" {
		return nil, fuego.BadRequestError{
			Detail: notification.ErrMissingID.Error(),
			Err:    notification.ErrMissingID,
		}
	}

	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{
			Detail: notification.ErrAccessDenied.Error(),
			Err:    notification.ErrAccessDenied,
		}
	}

	// Use org from context (already verified by AuthMiddleware via Better Auth).
	// Query param id must match the authenticated org to prevent cross-org access.
	ctxOrg := utils.GetOrganizationID(r)
	if ctxOrg == uuid.Nil {
		return nil, fuego.ForbiddenError{
			Detail: notification.ErrUserDoesNotBelongToOrganization.Error(),
			Err:    notification.ErrUserDoesNotBelongToOrganization,
		}
	}
	if ctxOrg.String() != id {
		return nil, fuego.ForbiddenError{
			Detail: notification.ErrUserDoesNotBelongToOrganization.Error(),
			Err:    notification.ErrUserDoesNotBelongToOrganization,
		}
	}
	orgID := id

	SMTPConfigs, err := c.service.GetSmtp(user.ID.String(), orgID)
	if err != nil {
		if errors.Is(err, notification.ErrSMTPConfigNotFound) {
			return smtpNotFoundResp(), nil
		}

		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
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
