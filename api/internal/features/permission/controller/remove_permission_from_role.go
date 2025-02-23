package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *PermissionController) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	var permission types.RemovePermissionFromRoleRequest

	if err := c.validator.ParseRequestBody(&permission, r.Body, &permission); err != nil {
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.RemovePermissionFromRole(permission.PermissionID, permission.RoleID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission removed from role successfully", nil)
}