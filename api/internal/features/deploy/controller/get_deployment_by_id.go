package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DeployController) GetDeploymentById(f fuego.ContextNoBody) (*types.DeploymentResponse, error) {
	deploymentID := f.PathParam("deployment_id")
	w, r := f.Response(), f.Request()
	data := deployRequestData(r, utils.GetUser(w, r))
	if data == "" && deploymentID != "" {
		data = fmt.Sprintf("deployment_id=%s", deploymentID)
	} else if deploymentID != "" {
		data = fmt.Sprintf("%s deployment_id=%s", data, deploymentID)
	}

	deployment, err := c.service.GetDeploymentById(deploymentID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("deploy: GetDeploymentById: %v", err), data)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.DeploymentResponse{
		Status:  "success",
		Message: "Deployment Retrieved successfully",
		Data:    deployment,
	}, nil
}
