package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetRoles returns all roles that are active in the database
// passing is_disabled as true will return all roles
func (c *RolesController) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := c.service.GetRoles()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, types.ErrFailedToGetRoles.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Roles fetched successfully", roles)
}

// GetRole returns a single role for the given id in the database
func (c *RolesController) GetRole(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	request := types.GetRoleRequest{
		ID: id,
	}

	if err := c.validator.ParseRequestBody(request, r.Body, &id); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
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
