package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (p *PermissionController)  AddPermissionToRole(w http.ResponseWriter, r *http.Request) {
	var permission types.AddPermissionToRoleRequest

	if err := p.validator.ParseRequestBody(&permission, r.Body, &permission); err != nil {
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := p.validator.ValidateRequest(permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := p.service.AddPermissionToRole(permission.PermissionID, permission.RoleID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission added to role successfully", nil)
}

