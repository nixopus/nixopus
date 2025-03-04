package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// CreateRole godoc
// @Summary Create a new role
// @Description Creates a new role in the application.
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param create_role body types.CreateRoleRequest true "Create role request"
// @Success 201 {object} types.Response "Success response with role"
// @Failure 400 {object} types.Response "Bad request"
// @Router /roles [post]
func (c *RolesController) CreateRole(w http.ResponseWriter, r *http.Request) {
	var role types.CreateRoleRequest

	if !c.parseAndValidate(w, r, &role) {
		return
	}

	err := c.service.CreateRole(&role)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Role created successfully", nil)
}
