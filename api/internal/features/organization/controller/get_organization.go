package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
)

func (c *OrganizationsController) GetOrganization(f fuego.ContextNoBody) (*types.OrganizationResponse, error) {
	id := f.QueryParam("id")
	if err := c.validator.ValidateID(id, types.ErrMissingOrganizationID); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	organization, err := c.service.GetOrganization(id)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.OrganizationResponse{
		Status:  "success",
		Message: "Organization fetched successfully",
		Data:    organization,
	}, nil
}
