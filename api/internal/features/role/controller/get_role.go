package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetRoles godoc
// @Summary Get all roles
// @Description Retrieves all roles in the database
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} types.Role "Success response with roles"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /roles [get]
func (c *RolesController) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := c.service.GetRoles()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, types.ErrFailedToGetRoles.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Roles fetched successfully", roles)
}

// GetRole godoc
// @Summary Get a role
// @Description Retrieves a role by its ID
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Role ID"
// @Success 200 {object} types.Role "Success response with role"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /roles/{id} [get]
func (c *RolesController) GetRole(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	request := types.GetRoleRequest{
		ID: id,
	}

	if err := c.validator.ValidateRequest(request); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	role, err := c.service.GetRole(id)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetRole.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Role fetched successfully", role)
}
