package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (u *UserController) GetIsOnboarded(s fuego.ContextNoBody) (*types.IsOnboardedResponse, error) {
	w, r := s.Response(), s.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		u.logUserDebug("GetIsOnboarded", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)
	u.logger.Log(logger.Info, "user: GetIsOnboarded", ctxStr)

	isOnboarded, err := u.service.IsOnboarded(user.ID.String())
	if err != nil {
		u.logger.Log(logger.Error, fmt.Sprintf("user: GetIsOnboarded: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
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
