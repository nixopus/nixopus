package controller

import (
	"net/http"
	
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *OrganizationsController) RemoveUserFromOrganization(w http.ResponseWriter, r *http.Request) {
	var user types.RemoveUserFromOrganizationRequest

	if err := c.validator.ParseRequestBody(r, r.Body, &user); err != nil {
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(user); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.RemoveUserFromOrganization(&user); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "User removed from organization successfully", nil)
}
