package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *DeployController) GetApplicationById(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	user := c.GetUser(w, r)

	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		return
	}

	application, err := c.service.GetApplicationById(id)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Application Retrieved successfully", application)
}
