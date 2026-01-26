package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (u *UserController) GetIsOnboarded(s fuego.ContextNoBody) (*types.IsOnboardedResponse, error) {
	w, r := s.Response(), s.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	u.logger.Log(logger.Info, "checking onboarding status", user.ID.String())

	isOnboarded, err := u.service.IsOnboarded(user.ID.String())
	if err != nil {
		u.logger.Log(logger.Error, err.Error(), user.ID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.IsOnboardedResponse{
		Status:  "success",
		Message: "Onboarding status fetched successfully",
		Data: types.IsOnboardedResponseData{
			IsOnboarded: isOnboarded,
		},
	}, nil
}
