package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// @Summary Delete SMTP configuration
// @Description Delete SMTP configuration
// @Tags notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param SMTPConfigs body notification.DeleteSMTPConfigRequest true "SMTP configuration"
// @Success 200 {object} types.Response "SMTP deleted successfully"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /notification/smtp [delete]
func (c *NotificationController) DeleteSmtp(w http.ResponseWriter, r *http.Request) {
	var SMTPConfigs notification.DeleteSMTPConfigRequest

	if !c.parseAndValidate(w, r, &SMTPConfigs) {
		return
	}

	err := c.service.DeleteSmtp(SMTPConfigs.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "SMTP deleted successfully", nil)
}
