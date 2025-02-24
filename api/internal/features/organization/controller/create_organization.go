package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// CreateOrganization godoc
// @Summary Create a new organization
// @Description Creates a new organization in the application.
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param create_organization body types.CreateOrganizationRequest true "Create organization request"
// @Success 200 {object} types.Response "Success response with organization"
// @Failure 400 {object} types.Response "Bad request"
// @Router /organization/create [post]
func (c *OrganizationsController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var organization types.CreateOrganizationRequest

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

	if err := c.service.CreateOrganization(&organization); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Organization created successfully", nil)
}
