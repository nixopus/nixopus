package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreateRole creates a new role in the database
func (c *RolesController) CreateRole(w http.ResponseWriter, r *http.Request) {
	var role types.CreateRoleRequest

	if err := c.validator.ParseRequestBody(r, r.Body, &role); err != nil {
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(role); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	err:= c.service.CreateRole(&role)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Role created successfully", nil)
}
