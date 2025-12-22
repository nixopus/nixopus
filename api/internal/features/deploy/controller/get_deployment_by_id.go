package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (c *DeployController) GetDeploymentById(f fuego.ContextNoBody) (*types.DeploymentResponse, error) {
	deploymentID := f.PathParam("deployment_id")

	deployment, err := c.service.GetDeploymentById(deploymentID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.DeploymentResponse{
		Status:  "success",
		Message: "Deployment Retrieved successfully",
		Data:    deployment,
	}, nil
}
