package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetUserDetails godoc
// @Summary Get user details endpoint
// @Description Retrieves the details of the current user.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.Response{data=types.User} "Success response with user details"
// @Failure 401 {object} types.Response "Unauthorized"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /user/details [get]
func (u *UserController) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	u.logger.Log(logger.Info, "getting user details", "")
	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		u.logger.Log(logger.Error, shared_types.ErrFailedToGetUserFromContext.Error(), shared_types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "User details retrieved successfully", user)
}
