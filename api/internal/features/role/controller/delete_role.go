package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// DeleteRole deletes a role from the database
func (c *RolesController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if err := c.validator.ValidateRequest(types.GetRoleRequest{ID: id}); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteRole(id); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Role deleted successfully", nil)
}
