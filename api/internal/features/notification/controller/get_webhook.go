package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/controller/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *NotificationController) GetWebhookConfig(f fuego.ContextNoBody) (*types.WebhookConfigResponse, error) {
	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	webhookType := f.PathParam("type")

	config, err := c.service.GetWebhookConfig(f, &notification.GetWebhookConfigRequest{Type: webhookType}, orgID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.WebhookConfigResponse{
		Status:  "success",
		Message: "Webhook config retrieved successfully",
		Data:    config,
	}, nil
}
