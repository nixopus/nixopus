package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *DeployController) IsNameAlreadyTaken(w http.ResponseWriter, r *http.Request) {
	var request types.IsNameAlreadyTakenRequest
	if !c.parseAndValidate(w, r, &request) {
		return
	}

	value, err := c.service.IsNameAlreadyTaken(request.Name)

	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendJSONResponse(w, "success", "", value)
}

func (c *DeployController) IsDomainAlreadyTaken(w http.ResponseWriter, r *http.Request) {
	var request types.IsDomainAlreadyTakenRequest
	if !c.parseAndValidate(w, r, &request) {
		return
	}

	is_taken, err := c.service.IsDomainAlreadyTaken(request.Domain)

	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	is_valid, err := c.service.IsDomainValid(request.Domain)

	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if is_taken || !is_valid {
		utils.SendJSONResponse(w, "success", "", false)
		return
	}

	utils.SendJSONResponse(w, "success", "", true)
}

func (c *DeployController) IsPortAlreadyTaken(w http.ResponseWriter, r *http.Request) {
	var request types.IsPortAlreadyTakenRequest
	if !c.parseAndValidate(w, r, &request) {
		return
	}

	value, err := c.service.IsPortAlreadyTaken(request.Port)

	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendJSONResponse(w, "success", "", value)
}
