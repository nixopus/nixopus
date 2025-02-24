package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *OrganizationsController) GetOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := c.validator.ValidateID(id, types.ErrMissingOrganizationID); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	users, err := c.service.GetOrganizationUsers(id)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Organization users fetched successfully", users)
}
