package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// CreatePermission godoc
// @Summary Create a new permission
// @Description Creates a new permission in the application.
// @Tags permission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param create_permission body types.CreatePermissionRequest true "Create permission request"
// @Success 201 {object} types.Response "Success response with permission"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /permissions [post]
func (p *PermissionController) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var permission types.CreatePermissionRequest

	if !p.parseAndValidate(w, r, &permission) {
		return
	}

	if err := p.service.CreatePermission(&permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission created successfully", nil)
}
