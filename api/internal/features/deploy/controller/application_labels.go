package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type UpdateLabelsRequest struct {
	Labels []string `json:"labels" validate:"required"`
}

func (c *DeployController) UpdateApplicationLabels(f fuego.ContextWithBody[UpdateLabelsRequest]) (*types.LabelsResponse, error) {
	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	applicationID := f.QueryParam("id")
	if applicationID == "" {
		c.logger.Log(logger.Error, "application id is required", "")
		return nil, fuego.BadRequestError{
			Detail: "application ID is required",
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	appID, err := uuid.Parse(applicationID)
	if err != nil {
		c.logger.Log(logger.Error, "invalid application id", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	err = c.service.UpdateApplicationLabels(appID, data.Labels, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.LabelsResponse{
		Status:  "success",
		Message: "Labels updated successfully",
		Data:    data.Labels,
	}, nil
}
