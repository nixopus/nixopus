package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// @Summary Get permissions by role
// @Description Get permissions by role
// @Tags permission
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id query string true "Role ID"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /permissions/roles/{id} [get]
func (p *PermissionController) GetPermissionsByRole(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	permissions, err := p.service.GetPermissionByRole(id)
	if err != nil {
		p.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, types.ErrFailedToGetPermission.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Permissions fetched successfully", permissions)
}
