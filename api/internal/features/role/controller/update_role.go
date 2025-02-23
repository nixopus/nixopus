package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// UpdateRole updates a role in the database
// Takes in four parameters: id and name, description, isDeleted is optional
func (c *RolesController) UpdateRole(w http.ResponseWriter, r *http.Request) {
	var role types.UpdateRoleRequest

	if err := c.validator.ParseRequestBody(&role, r.Body, &role); err != nil {
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(role); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateRole(role.ID, role); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Role updated successfully", nil)
}