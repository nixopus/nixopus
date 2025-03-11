package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// HandleDeploy handles the request to create a new deployment
// It requires a valid user and a valid deployment request in the request body
// It returns the created deployment in the response body with a status code of 200
// If something fails, it returns an error response with a status code of 500
func (c *DeployController) HandleDeploy(w http.ResponseWriter, r *http.Request) {
	c.logger.Log(logger.Info, "deploying", "")
	var data types.CreateDeploymentRequest

	if !c.parseAndValidate(w, r, &data) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	application, err := c.service.CreateDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to create deployment", err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Deployment created successfully", application)
}
