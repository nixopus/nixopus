package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (u *UserController) UpdateAvatar(s fuego.ContextWithBody[types.UpdateAvatarRequest]) (*types.MessageResponse, error) {
	req, err := s.Body()
	if err != nil {
		u.logUserDebug("UpdateAvatar", fmt.Sprintf("parse body: %v", err), "")
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
		u.logUserDebug("UpdateAvatar", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := fmt.Sprintf("%s avatar_payload_len=%d", userRequestData(r, user), len(req.AvatarData))
	u.logger.Log(logger.Info, "user: UpdateAvatar", ctxStr)

	err = u.service.UpdateAvatar(s.Request().Context(), user.ID.String(), &req)
	if err != nil {
		u.logger.Log(logger.Error, fmt.Sprintf("user: UpdateAvatar: %v", err), userRequestData(r, user))
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
