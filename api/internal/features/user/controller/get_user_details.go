package controller

import (
	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (u *UserController) GetUserDetails(s fuego.ContextNoBody) (*types.UserResponse, error) {
	w, r := s.Response(), s.Request()

	user := utils.GetUser(w, r)

	if user == nil {
		u.logUserDebug("GetUserDetails", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)
	u.logger.Log(logger.Info, "user: GetUserDetails", ctxStr)

	return &types.UserResponse{
		Status:  "success",
		Message: "User details fetched successfully",
		Data:    user,
	}, nil
}
