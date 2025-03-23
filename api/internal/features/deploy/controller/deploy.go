package controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)


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

func (c *DeployController) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	c.logger.Log(logger.Info, "deleting application", "")
	var data types.DeleteDeploymentRequest

	if !c.parseAndValidate(w, r, &data) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	err := c.service.DeleteDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to delete application", err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Application deleted successfully", nil)
}

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

func (c *DeployController) HandleRollback(w http.ResponseWriter, r *http.Request) {
	c.logger.Log(logger.Info, "rolling back application", "")
	var data types.RollbackDeploymentRequest

	if !c.parseAndValidate(w, r, &data) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	err := c.service.RollbackDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to rollback application", err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Application rolled back successfully", nil)
}

func (c *DeployController) HandleRestart(w http.ResponseWriter, r *http.Request) {
	c.logger.Log(logger.Info, "restarting application", "")
	var data types.RestartDeploymentRequest

	if !c.parseAndValidate(w, r, &data) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	err := c.service.RestartDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to restart application", err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Application restarted successfully", nil)
}
