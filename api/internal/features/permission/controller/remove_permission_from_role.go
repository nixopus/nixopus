package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// @Summary Remove permission from role
// @Description Remove permission from role
// @Tags permission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body types.RemovePermissionFromRoleRequest true "Remove permission from role request"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /permissions/roles/remove [post]
func (c *PermissionController) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	var permission types.RemovePermissionFromRoleRequest

	if err := c.validator.ParseRequestBody(permission, r.Body, &permission); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(permission); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.RemovePermissionFromRole(permission.PermissionID, permission.RoleID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission removed from role successfully", nil)
}
