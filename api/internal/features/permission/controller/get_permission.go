package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetPermissions godoc
// @Summary Get all permissions
// @Description Retrieves all permissions.
// @Tags permission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} types.Permission "Success response with permissions"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /permissions [get]
func (c *PermissionController) GetPermissions(w http.ResponseWriter, r *http.Request) {
	permission, err := c.storage.GetPermissions()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, types.ErrFailedToGetPermissions.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Permissions fetched successfully", permission)
}

// GetPermission godoc
// @Summary Get a permission
// @Description Retrieves a permission by its ID.
// @Tags permission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Permission ID"
// @Success 200 {object} types.Permission "Success response with permission"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /permissions/{id} [get]
func (c *PermissionController) GetPermission(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	permission, err := c.storage.GetPermission(id)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetPermission.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Permission fetched successfully", permission)
}
