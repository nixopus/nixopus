package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// UpdatePermission godoc
// @Summary Update a permission
// @Description Updates a permission
// @Tags permission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param update_permission body types.UpdatePermissionRequest true "Update permission request"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /permissions/update [post]
func (c *PermissionController) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	var permission types.UpdatePermissionRequest

	if err := c.validator.ParseRequestBody(&permission, r.Body, &permission); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(permission); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.UpdatePermission(&permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission updated successfully", nil)
}
