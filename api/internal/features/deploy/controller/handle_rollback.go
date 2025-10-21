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

func (c *DeployController) HandleRollback(f fuego.ContextWithBody[types.RollbackDeploymentRequest]) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "starting application rollback process", "")

	data, err := f.Body()
	if err != nil {
		if err == io.EOF {
			c.logger.Log(logger.Error, "empty request body received", "id is required for rollback")
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

	c.logger.Log(logger.Info, "attempting to rollback application", "id: "+data.ID.String()+", user_id: "+user.ID.String())

	err = c.taskService.RollbackDeployment(&data, user.ID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to rollback application", "id: "+data.ID.String()+", error: "+err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "application rolled back successfully", "id: "+data.ID.String())
	return &shared_types.Response{
		Status:  "success",
		Message: "Application rolled back successfully",
		Data:    nil,
	}, nil
}
