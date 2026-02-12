package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *DeployController) GetApplicationById(f fuego.ContextNoBody) (*types.ApplicationResponse, error) {
	id := f.QueryParam("id")

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

	application, err := c.service.GetApplicationById(id, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		status := http.StatusInternalServerError
		if err.Error() == "application not found" {
			status = http.StatusNotFound
		}
		return nil, fuego.HTTPError{
			Err:    err,
			Status: status,
		}
	}

	return &types.ApplicationResponse{
		Status:  "success",
		Message: "Application Retrieved successfully",
		Data:    application,
	}, nil
}
