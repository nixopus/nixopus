package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DeployController) GetApplicationById(f fuego.ContextNoBody) (*types.ApplicationResponse, error) {
	id := f.QueryParam("id")

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

	application, err := c.service.GetApplicationById(id, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		if err.Error() == "application not found" {
			return nil, fuego.NotFoundError{
				Detail: err.Error(),
				Err:    err,
			}
		}
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ApplicationResponse{
		Status:  "success",
		Message: "Application Retrieved successfully",
		Data:    application,
	}, nil
}
