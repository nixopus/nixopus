package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetOrganizationSettings retrieves organization settings
func (c *OrganizationsController) GetOrganizationSettings(s fuego.ContextNoBody) (*types.OrganizationSettingsResponse, error) {
	_, r := s.Response(), s.Request()
	orgID := utils.GetOrganizationID(r)

	if orgID == uuid.Nil {
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingOrganizationID,
			Status: http.StatusBadRequest,
		}
	}

	settings, err := c.service.GetOrganizationSettings(orgID.String())
	if err != nil {
		c.logger.Log(logger.Error, "failed to get organization settings", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.OrganizationSettingsResponse{
		Status:  "success",
		Message: "Organization settings fetched successfully",
		Data:    settings,
	}, nil
}

// UpdateOrganizationSettings updates organization settings with the provided data
func (c *OrganizationsController) UpdateOrganizationSettings(s fuego.ContextWithBody[shared_types.OrganizationSettingsData]) (*types.OrganizationSettingsResponse, error) {
	_, r := s.Response(), s.Request()
	orgID := utils.GetOrganizationID(r)

	if orgID == uuid.Nil {
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingOrganizationID,
			Status: http.StatusBadRequest,
		}
	}

	req, err := s.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	settings, err := c.service.UpdateOrganizationSettings(orgID.String(), req)
	if err != nil {
		c.logger.Log(logger.Error, "failed to update organization settings", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.OrganizationSettingsResponse{
		Status:  "success",
		Message: "Organization settings updated successfully",
		Data:    settings,
	}, nil
}
