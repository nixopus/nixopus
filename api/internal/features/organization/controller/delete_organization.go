package controller

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *OrganizationsController) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	var organization types.DeleteOrganizationRequest

	if err := c.validator.ParseRequestBody(r, r.Body, &organization); err != nil {
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(organization); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	organizationID, err := uuid.Parse(organization.ID)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteOrganization(organizationID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Organization deleted successfully", nil)
}
