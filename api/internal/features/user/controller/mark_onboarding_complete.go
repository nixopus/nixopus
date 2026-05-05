package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// MarkOnboardingComplete marks the authenticated user's onboarding as complete.
func (u *UserController) MarkOnboardingComplete(s fuego.ContextNoBody) (*types.MarkOnboardingCompleteResponse, error) {
	w, r := s.Response(), s.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		u.logUserDebug("MarkOnboardingComplete", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)
	u.logger.Log(logger.Info, "user: MarkOnboardingComplete", ctxStr)

	err := u.service.MarkOnboardingComplete(user.ID.String())
	if err != nil {
		u.logger.Log(logger.Error, fmt.Sprintf("user: MarkOnboardingComplete: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	if u.cache != nil {
		_ = u.cache.InvalidateUserByID(u.ctx, user.ID.String())
	}

	return &types.MarkOnboardingCompleteResponse{
		Data: types.IsOnboardedResponseData{
			IsOnboarded: true,
		},
	}, nil
}
