package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type UpdateLabelsRequest struct {
	Labels []string `json:"labels" validate:"required"`
}

func (c *DeployController) UpdateApplicationLabels(f fuego.ContextWithBody[UpdateLabelsRequest]) (*types.Response, error) {
	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	applicationID := f.QueryParam("id")
	if applicationID == "" {
		c.logger.Log(logger.Error, "application id is required", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	appID, err := uuid.Parse(applicationID)
	if err != nil {
		c.logger.Log(logger.Error, "invalid application id", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	err = c.service.UpdateApplicationLabels(appID, data.Labels, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.Response{
		Status:  "success",
		Message: "Labels updated successfully",
		Data:    data.Labels,
	}, nil
}
