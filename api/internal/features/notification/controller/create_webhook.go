package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *NotificationController) CreateWebhookConfig(f fuego.ContextWithBody[notification.CreateWebhookConfigRequest]) (*types.WebhookConfigResponse, error) {
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

	config, err := c.service.CreateWebhookConfig(f, &req, user.ID, orgID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.WebhookConfigResponse{
		Status:  "success",
		Message: "Webhook config created successfully",
		Data:    config,
	}, nil
}
