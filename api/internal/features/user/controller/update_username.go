package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
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

	if !u.parseAndValidate(w, r, &req) {
		return
	}

	user := u.GetUser(w, r)
	if user == nil {
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
