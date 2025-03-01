package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// @Summary Add permission to role
// @Description Add permission to role
// @Tags permission
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param request body types.AddPermissionToRoleRequest true "Add permission to role request"
// @Success 201 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /permissions/roles [post]
func (p *PermissionController) AddPermissionToRole(w http.ResponseWriter, r *http.Request) {
	var permission types.AddPermissionToRoleRequest

	if err := p.validator.ParseRequestBody(&permission, r.Body, &permission); err != nil {
		p.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := p.validator.ValidateRequest(permission); err != nil {
		p.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := p.service.AddPermissionToRole(permission.PermissionID, permission.RoleID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission added to role successfully", nil)
}
