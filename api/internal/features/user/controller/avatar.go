package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (u *UserController) UpdateAvatar(s fuego.ContextWithBody[types.UpdateAvatarRequest]) (*types.MessageResponse, error) {
	u.logger.Log(logger.Info, "updating user avatar", "")

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

	err = u.service.UpdateAvatar(s.Request().Context(), user.ID.String(), &req)
	if err != nil {
		u.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	u.cache.InvalidateUserByID(u.ctx, user.ID.String())

	return &types.MessageResponse{
		Status:  "success",
		Message: "Avatar updated successfully",
	}, nil
}
