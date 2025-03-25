package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// @Summary Get notification preferences
// @Description Retrieves the notification preferences for the current user
// @Tags notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.Response{data=notification.GetPreferencesResponse} "Success response with preferences"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /notification/preferences [get]
func (c *NotificationController) GetPreferences(w http.ResponseWriter, r *http.Request) {
	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	preferences, err := c.service.GetPreferences(user.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Notification preferences", preferences)
}
