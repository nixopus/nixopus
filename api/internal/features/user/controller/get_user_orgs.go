package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetUserOrganizations godoc
// @Summary Get user organizations
// @Description Retrieves the organizations for the current user.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} types.Organization "Success response with organizations"
// @Failure 401 {object} types.Response "Unauthorized"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /user/organizations [get]
func (u *UserController) GetUserOrganizations(w http.ResponseWriter, r *http.Request) {
	u.logger.Log(logger.Info, "getting user organizations", "")

	user := u.GetUser(w, r)
	if user == nil {
		return
	}

	organizations, err := u.service.GetUserOrganizations(user.ID.String())

	if err != nil {
		u.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "User organizations retrieved successfully", organizations)
}
