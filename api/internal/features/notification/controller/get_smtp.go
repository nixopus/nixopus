package controller

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"net/http"
)

// @Summary Get SMTP configuration
// @Description Get SMTP configuration
// @Tags notification
// @Accept json
// @Produce json
// @Success 200 {object} types.Response "SMTP configuration"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /notification/get-smtp [get]
func (c *NotificationController) GetSmtp(w http.ResponseWriter, r *http.Request) {
	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	SMTPConfigs, err := c.service.GetSmtp(user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "SMTP configuration", SMTPConfigs)
}
