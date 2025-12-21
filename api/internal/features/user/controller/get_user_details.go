package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (u *UserController) GetUserDetails(s fuego.ContextNoBody) (*types.UserResponse, error) {
	w, r := s.Response(), s.Request()

	user := utils.GetUser(w, r)

	u.logger.Log(logger.Info, "getting user details", "")

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	return &types.UserResponse{
		Status:  "success",
		Message: "User details fetched successfully",
		Data:    user,
	}, nil
}
