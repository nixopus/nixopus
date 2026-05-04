package controller

import (
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DeployController) HandleRollback(f fuego.ContextWithBody[types.RollbackDeploymentRequest]) (*types.MessageResponse, error) {
	c.logger.Log(logger.Info, "deploy: starting application rollback process", "")

	data, err := f.Body()
	if err != nil {
		if err == io.EOF {
			c.logger.Log(logger.Error, "deploy: empty request body received", "id is required for rollback")
			return nil, fuego.BadRequestError{
				Detail: types.ErrMissingID.Error(),
				Err:    types.ErrMissingID,
			}
		}
		c.logger.Log(logger.Error, "deploy: failed to read request body", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	c.logger.Log(logger.Info, "deploy: request body parsed successfully", "id: "+data.ID.String())

	if err := c.validator.ValidateRequest(&data); err != nil {
		c.logger.Log(logger.Error, "deploy: request validation failed", "id: "+data.ID.String()+", error: "+err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "deploy: user authentication failed", "id: "+data.ID.String())
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "deploy: organization not found", "id: "+data.ID.String())
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	c.logger.Log(logger.Info, "deploy: attempting to rollback application", "id: "+data.ID.String()+", user_id: "+user.ID.String())

	err = c.taskService.RollbackDeployment(&data, user.ID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "deploy: failed to rollback application", "id: "+data.ID.String()+", error: "+err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "deploy: application rolled back successfully", "id: "+data.ID.String())
	return &types.MessageResponse{
		Status:  "success",
		Message: "Application rolled back successfully",
	}, nil
}
