package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *NotificationController) GetWebhookConfig(f fuego.ContextNoBody) (*shared_types.Response, error) {
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

	return &shared_types.Response{
		Status:  "success",
		Message: "Webhook config retrieved successfully",
		Data:    config,
	}, nil
}
