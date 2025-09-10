package controller

import (
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *DeployController) UpdateApplication(f fuego.ContextWithBody[types.UpdateDeploymentRequest]) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "starting application update process", "")

	data, err := f.Body()
	if err != nil {
		if err == io.EOF {
			c.logger.Log(logger.Error, "empty request body received", "id is required for update")
			return nil, fuego.HTTPError{
				Err:    types.ErrMissingID,
				Status: http.StatusBadRequest,
			}
		}
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	c.logger.Log(logger.Info, "request body parsed successfully", "id: "+data.ID.String())

	if err := c.validator.ValidateRequest(&data); err != nil {
		c.logger.Log(logger.Error, "request validation failed", "id: "+data.ID.String()+", error: "+err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user authentication failed", "id: "+data.ID.String())
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "id: "+data.ID.String())
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	c.logger.Log(logger.Info, "attempting to update application", "id: "+data.ID.String()+", user_id: "+user.ID.String())

	// application, err := c.service.UpdateDeployment(&data, user.ID, organizationID)
	// if err != nil {
	// 	c.logger.Log(logger.Error, "failed to update application", "id: "+data.ID.String()+", error: "+err.Error())
	// 	return nil, fuego.HTTPError{
	// 		Err:    err,
	// 		Status: http.StatusInternalServerError,
	// 	}
	// }

	application, err := c.taskService.UpdateDeployment(&data, user.ID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to create deployment", "name: "+data.Name+", error: "+err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "application updated successfully", "id: "+data.ID.String())
	return &shared_types.Response{
		Status:  "success",
		Message: "Application updated successfully",
		Data:    application,
	}, nil
}
