package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (u *UserController) GetUserDetails(s fuego.ContextNoBody) (*shared_types.Response, error) {
	w, r := s.Response(), s.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "User details fetched successfully",
		Data:    user,
	}, nil
}
