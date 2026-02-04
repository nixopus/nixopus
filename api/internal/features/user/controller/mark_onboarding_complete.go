package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// MarkOnboardingComplete marks the authenticated user's onboarding as complete.
func (u *UserController) MarkOnboardingComplete(s fuego.ContextNoBody) (*types.MarkOnboardingCompleteResponse, error) {
	w, r := s.Response(), s.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	u.logger.Log(logger.Info, "marking onboarding as complete", user.ID.String())

	err := u.service.MarkOnboardingComplete(user.ID.String())
	if err != nil {
		u.logger.Log(logger.Error, err.Error(), user.ID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MarkOnboardingCompleteResponse{
		Data: types.IsOnboardedResponseData{
			IsOnboarded: true,
		},
	}, nil
}
