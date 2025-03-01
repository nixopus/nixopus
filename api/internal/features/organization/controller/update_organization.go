package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// UpdateOrganization godoc
// @Summary Update an organization
// @Description Updates an organization
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param update_organization body types.UpdateOrganizationRequest true "Update organization request"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /organizations [put]
func (c *OrganizationsController) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	var organization types.UpdateOrganizationRequest

	c.logger.Log(logger.Info, "updating organization", organization.ID)

	if err := c.validator.ParseRequestBody(r, r.Body, &organization); err != nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(organization); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateOrganization(&organization); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Organization updated successfully", nil)
}
