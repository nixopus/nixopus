package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (u *UserController) UpdateUserName(s fuego.ContextWithBody[types.UpdateUserNameRequest]) (*types.UpdateUsernameResponse, error) {
	req, err := s.Body()
	if err != nil {
		u.logUserDebug("UpdateUserName", fmt.Sprintf("parse body: %v", err), "")
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	w, r := s.Response(), s.Request()

	if err := u.parseAndValidate(w, r, &req); err != nil {
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		u.logUserDebug("UpdateUserName", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)
	u.logger.Log(logger.Info, "user: UpdateUserName", ctxStr)

	err = u.service.UpdateUsername(user.ID.String(), &req)

	if err != nil {
		u.logger.Log(logger.Error, fmt.Sprintf("user: UpdateUserName: %v", err), ctxStr)
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
