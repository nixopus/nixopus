package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// UpdateUserName godoc
// @Summary Update user name endpoint
// @Description Updates the user's name based on the provided information.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param updateUserName body types.UpdateUserNameRequest true "Update user name request"
// @Success 200 {object} types.Response "User name updated successfully"
// @Failure 400 {object} types.Response "Failed to decode or validate request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /user/update-username [put]
func (u *UserController) UpdateUserName(w http.ResponseWriter, r *http.Request) {
	u.logger.Log(logger.Info, "updating user name", "")

	var req types.UpdateUserNameRequest

	if err := u.validator.ParseRequestBody(r, r.Body, &req); err != nil {
		u.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		u.logger.Log(logger.Error, shared_types.ErrFailedToGetUserFromContext.Error(), shared_types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	err := u.service.UpdateUsername(user.ID.String(), &req)

	if err != nil {
		u.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "User name updated successfully", req.Name)
}

