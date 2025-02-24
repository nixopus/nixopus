package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// DeleteRole godoc
// @Summary Delete a role
// @Description Deletes a role with the given id.
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Role ID"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /roles/delete [delete]
func (c *RolesController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if err := c.validator.ValidateRequest(types.GetRoleRequest{ID: id}); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteRole(id); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Role deleted successfully", nil)
}
