package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// @Summary Update SMTP configuration
// @Description Update SMTP configuration
// @Tags notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param SMTPConfigs body notification.UpdateSMTPConfigRequest true "SMTP configuration"
// @Success 200 {object} types.Response "SMTP updated successfully"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /notification/smtp [put]
func (c *NotificationController) UpdateSmtp(w http.ResponseWriter, r *http.Request) {
	var SMTPConfigs notification.UpdateSMTPConfigRequest

	if !c.parseAndValidate(w, r, &SMTPConfigs) {
		return
	}

	err := c.service.UpdateSmtp(SMTPConfigs)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "SMTP updated successfully", nil)
}
