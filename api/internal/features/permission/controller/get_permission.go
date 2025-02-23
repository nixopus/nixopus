package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *PermissionController) GetPermissions(w http.ResponseWriter, r *http.Request) {
	permission, err := c.storage.GetPermissions()
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetPermissions.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Permissions fetched successfully", permission)
}

func (c *PermissionController) GetPermission(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	request := types.GetPermissionRequest{
		ID: id,
	}
	if err := c.validator.ParseRequestBody(request, r.Body, &id); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	permission, err := c.storage.GetPermission(id)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetPermission.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Permission fetched successfully", permission)
}
