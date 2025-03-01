package controller

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"net/http"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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
	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		c.logger.Log(logger.Error, shared_types.ErrFailedToGetUserFromContext.Error(), shared_types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
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