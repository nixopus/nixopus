package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// @Summary Add SMTP configuration
// @Description Add SMTP configuration
// @Tags notification
// @Accept json
// @Produce json
// @Param SMTPConfigs body notification.CreateSMTPConfigRequest true "SMTP configuration"
// @Success 200 {object} types.Response "SMTP added successfully"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /notification/add-smtp [post]
func (c *NotificationController) AddSmtp(w http.ResponseWriter, r *http.Request) {
	var SMTPConfigs notification.CreateSMTPConfigRequest
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

	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		c.logger.Log(logger.Error, shared_types.ErrFailedToGetUserFromContext.Error(), shared_types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	err := c.service.AddSmtp(SMTPConfigs,user.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "SMTP added successfully", nil)
}
