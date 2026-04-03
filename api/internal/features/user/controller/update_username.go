package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (u *UserController) UpdateUserName(s fuego.ContextWithBody[types.UpdateUserNameRequest]) (*types.UpdateUsernameResponse, error) {
	u.logger.Log(logger.Info, "updating user name", "")

	req, err := s.Body()
	if err != nil {
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	w, r := s.Response(), s.Request()

	if !u.parseAndValidate(w, r, &req) {
		return nil, fuego.BadRequestError{
			Detail: "validation failed",
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	err = u.service.UpdateUsername(user.ID.String(), &req)

	if err != nil {
		u.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	u.cache.InvalidateUserByID(u.ctx, user.ID.String())

	return &types.UpdateUsernameResponse{
		Status:  "success",
		Message: "Username updated successfully",
		Data:    types.UpdateUsernameResponseData{Name: req.Name},
	}, nil
}
