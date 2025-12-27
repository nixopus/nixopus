package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (u *UserController) GetUserOrganizations(s fuego.ContextNoBody) (*types.UserOrganizationsListResponse, error) {
	w, r := s.Response(), s.Request()
	u.logger.Log(logger.Info, "getting user organizations", "")

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizations, err := u.service.GetUserOrganizations(user.ID.String())

	if err != nil {
		u.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UserOrganizationsListResponse{
		Status:  "success",
		Message: "User organizations fetched successfully",
		Data:    organizations,
	}, nil
}
