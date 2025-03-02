package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// @Summary Delete SMTP configuration
// @Description Delete SMTP configuration
// @Tags notification
// @Accept json
// @Produce json
// @Param SMTPConfigs body notification.DeleteSMTPConfigRequest true "SMTP configuration"
// @Success 200 {object} types.Response "SMTP deleted successfully"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /notification/delete-smtp [post]
func (c *NotificationController) DeleteSmtp(w http.ResponseWriter, r *http.Request) {
	var SMTPConfigs notification.DeleteSMTPConfigRequest
	if err := c.validator.ParseRequestBody(r, r.Body, &SMTPConfigs); err != nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(SMTPConfigs); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
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
