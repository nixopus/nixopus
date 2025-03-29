package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *DeployController) HandleDeploy(f fuego.ContextWithBody[types.CreateDeploymentRequest]) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "deploying", "")

	data, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	// if !c.parseAndValidate(w, r, &data) {
	// 	return nil, fuego.HTTPError{
	// 		Err:    nil,
	// 		Status: http.StatusBadRequest,
	// 	}
	// }

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	application, err := c.service.CreateDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to create deployment", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Deployment created successfully",
		Data:    application,
	}, nil
}

func (c *DeployController) UpdateApplication(f fuego.ContextWithBody[types.UpdateDeploymentRequest]) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "updating application", "")

	data, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	if !c.parseAndValidate(w, r, &data) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	application, err := c.service.UpdateDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to update application", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Application updated successfully",
		Data:    application,
	}, nil
}

func (c *DeployController) DeleteApplication(f fuego.ContextWithBody[types.DeleteDeploymentRequest]) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "deleting application", "")

	data, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	if !c.parseAndValidate(w, r, &data) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	err = c.service.DeleteDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to delete application", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Application deleted successfully",
		Data:    nil,
	}, nil
}

func (c *DeployController) ReDeployApplication(f fuego.ContextWithBody[types.ReDeployApplicationRequest]) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "redeploying application", "")

	data, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	if !c.parseAndValidate(w, r, &data) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	application, err := c.service.ReDeployApplication(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to redeploy application", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Application redeployed successfully",
		Data:    application,
	}, nil
}

func (c *DeployController) GetDeploymentById(f fuego.ContextNoBody) (*shared_types.Response, error) {
	deploymentID := f.QueryParam("deployment_id")

	deployment, err := c.service.GetDeploymentById(deploymentID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Deployment Retrieved successfully",
		Data:    deployment,
	}, nil
}

func (c *DeployController) HandleRollback(f fuego.ContextWithBody[types.RollbackDeploymentRequest]) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "rolling back application", "")

	data, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	if !c.parseAndValidate(w, r, &data) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	err = c.service.RollbackDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to rollback application", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Application rolled back successfully",
		Data:    nil,
	}, nil
}

func (c *DeployController) HandleRestart(f fuego.ContextWithBody[types.RestartDeploymentRequest]) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "restarting application", "")

	data, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	if !c.parseAndValidate(w, r, &data) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	err = c.service.RestartDeployment(&data, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to restart application", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Application restarted successfully",
		Data:    nil,
	}, nil
}
