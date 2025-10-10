package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type GetOrganizationUsersRequest struct {
	ID string `json:"id"`
}

func (c *OrganizationsController) GetOrganizationUsers(f fuego.ContextWithBody[GetOrganizationUsersRequest]) (*shared_types.Response, error) {
	_, r := f.Response(), f.Request()
	id := r.URL.Query().Get("id")
	if err := c.validator.ValidateID(id, types.ErrMissingOrganizationID); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}
	users, err := c.service.GetOrganizationUsersWithRoles(id)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Users fetched successfully",
		Data:    users,
	}, nil
}
