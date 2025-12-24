package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (u *UserController) UpdateUserName(s fuego.ContextWithBody[types.UpdateUserNameRequest]) (*types.UpdateUsernameResponse, error) {
	u.logger.Log(logger.Info, "updating user name", "")

	req, err := s.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := s.Response(), s.Request()

	if !u.parseAndValidate(w, r, &req) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	err = u.service.UpdateUsername(user.ID.String(), &req)

	if err != nil {
		u.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	u.cache.InvalidateUser(u.ctx, user.ID.String())

	return &types.UpdateUsernameResponse{
		Status:  "success",
		Message: "Username updated successfully",
		Data:    types.UpdateUsernameResponseData{Name: req.Name},
	}, nil
}
