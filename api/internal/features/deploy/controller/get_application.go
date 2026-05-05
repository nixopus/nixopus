package controller

import (
	"fmt"
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
		c.logger.Log(logger.Error, "deploy: user not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "deploy: organization not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	if id != "" {
		if _, err := uuid.Parse(id); err != nil {
			return nil, fuego.BadRequestError{
				Detail: "invalid application ID format",
			}
		}
	}

	data := deployRequestData(f.Request(), user)
	if id != "" {
		if data == "" {
			data = fmt.Sprintf("id=%s", id)
		} else {
			data = fmt.Sprintf("%s id=%s", data, id)
		}
	}

	application, err := c.service.GetApplicationById(id, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("deploy: GetApplicationById: %v", err), data)
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
