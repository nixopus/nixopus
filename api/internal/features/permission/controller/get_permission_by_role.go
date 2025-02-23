package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (p *PermissionController) GetPermissionsByRole(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	request := types.GetPermissionRequest{
		ID: id,
	}
	if err := p.validator.ParseRequestBody(request, r.Body, &id); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	permissions,err:= p.service.GetPermissionByRole(id)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetPermission.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Permissions fetched successfully", permissions)
}
