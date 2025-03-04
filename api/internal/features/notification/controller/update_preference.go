package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *NotificationController) UpdatePreference(w http.ResponseWriter, r *http.Request) {
	var prefRequest notification.UpdatePreferenceRequest
	if !c.parseAndValidate(w, r, &prefRequest) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	err := c.service.UpdatePreference(prefRequest, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Notification preferences updated successfully", nil)
}
