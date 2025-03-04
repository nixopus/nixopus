package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// UpdateRole godoc
// @Summary Update a role
// @Description Updates a role with the given id.
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param update_role body types.UpdateRoleRequest true "Update role request"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /roles/update [post]
func (c *RolesController) UpdateRole(w http.ResponseWriter, r *http.Request) {
	var role types.UpdateRoleRequest

	if !c.parseAndValidate(w, r, &role) {
		return
	}

	if err := c.service.UpdateRole(role.ID, role); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Role updated successfully", nil)
}
