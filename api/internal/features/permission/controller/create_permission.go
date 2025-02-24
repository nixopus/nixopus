package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (p *PermissionController) CreatePermission(w http.ResponseWriter, r *http.Request) {

	var permission types.CreatePermissionRequest

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

	if err := p.service.CreatePermission(&permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission created successfully", nil)
}
