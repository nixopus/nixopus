package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *NotificationController) GetPreferences(f fuego.ContextNoBody) (*types.PreferencesResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logNotificationDebug("GetPreferences", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := fmt.Sprintf("user_id=%s", user.ID)
	c.logger.Log(logger.Info, "notification: GetPreferences", ctxStr)

	preferences, err := c.service.GetPreferences(user.ID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("notification: GetPreferences: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.PreferencesResponse{
		Status:  "success",
		Message: "Preferences fetched successfully",
		Data:    preferences,
	}, nil
}
