package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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
	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		u.logger.Log(logger.Error, shared_types.ErrFailedToGetUserFromContext.Error(), shared_types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
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