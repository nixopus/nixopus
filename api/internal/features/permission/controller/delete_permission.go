package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// DeletePermission godoc
// @Summary Delete a permission
// @Description Deletes a permission with the given id.
// @Tags permission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param delete_permission body types.DeletePermissionRequest true "Delete permission request"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /permissions/delete [post]
func (c *PermissionController) DeletePermission(w http.ResponseWriter, r *http.Request) {
	var permission types.DeletePermissionRequest

	if !c.parseAndValidate(w, r, &permission) {
		return
	}

	if err := c.service.DeletePermission(permission.ID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission deleted successfully", nil)
}
