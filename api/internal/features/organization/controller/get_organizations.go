package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
)

func (c *OrganizationsController) GetOrganizations(f fuego.ContextNoBody) (*types.ListOrganizationsResponse, error) {
	organizations, err := c.service.GetOrganizations()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ListOrganizationsResponse{
		Status:  "success",
		Message: "Organizations fetched successfully",
		Data:    organizations,
	}, nil
}
