package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *PermissionController) DeletePermission(w http.ResponseWriter, r *http.Request) {
	var permission types.DeletePermissionRequest

	if err := c.validator.ParseRequestBody(&permission, r.Body, &permission); err != nil {
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.DeletePermission(permission.ID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission deleted successfully", nil)
}