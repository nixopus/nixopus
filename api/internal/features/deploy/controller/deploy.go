package controller

import (
	"net/http"

	"github.com/gorilla/mux"
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

// UpdateApplication updates an existing application deployment
//
// This endpoint is accessible by the authenticated user.
//
// @Summary Update an application deployment
// @Description Updates an existing application deployment in the application.
// @Tags deploy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deployment body types.UpdateDeploymentRequest true "Deployment update request"
// @Success 200 {object} types.UpdateDeploymentResponse "Success response with updated deployment"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /deploy [patch]
func (c *DeployController) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	c.logger.Log(logger.Info, "updating application", "")
	var data types.UpdateDeploymentRequest

	if !c.parseAndValidate(w, r, &data) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	application, err := c.service.UpdateDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to update application", err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Application updated successfully", application)
}

// func (c *DeployController) DeleteApplication(w http.ResponseWriter, r *http.Request) {
// 	c.logger.Log(logger.Info, "deleting application", "")
// 	var data types.DeleteDeploymentRequest

// 	if !c.parseAndValidate(w, r, &data) {
// 		return
// 	}

// 	user := c.GetUser(w, r)

// 	if user == nil {
// 		return
// 	}

// 	application, err := c.service.DeleteDeployment(&data, user.ID)
// 	if err != nil {
// 		c.logger.Log(logger.Error, "failed to delete application", err.Error())
// 		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	utils.SendJSONResponse(w, "success", "Application deleted successfully", application)
// }

func (c *DeployController) ReDeployApplication(w http.ResponseWriter, r *http.Request) {
	c.logger.Log(logger.Info, "redeploying application", "")
	var data types.ReDeployApplicationRequest

	if !c.parseAndValidate(w, r, &data) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	application, err := c.service.ReDeployApplication(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to redeploy application", err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Application redeployed successfully", application)
}

func (c *DeployController) GetDeploymentById(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deploymentID := vars["deployment_id"]
    
	deployment, err := c.service.GetDeploymentById(deploymentID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	utils.SendJSONResponse(w, "success", "Deployment Retrieved successfully", deployment)
}