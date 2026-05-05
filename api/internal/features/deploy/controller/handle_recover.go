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

func (c *DeployController) HandleRecover(f fuego.ContextWithBody[types.RecoverRequest]) (*types.RecoverResponse, error) {
	c.logger.Log(logger.Info, "deploy: starting application recovery process", "")

	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "deploy: failed to read request body", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "deploy: user authentication failed during recovery", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "deploy: organization not found during recovery", "")
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	c.logger.Log(logger.Info, "deploy: recovering applications", "org_id: "+organizationID.String())

	result, err := c.taskService.RecoverApplications(f.Request().Context(), organizationID, data.ApplicationID)
	if err != nil {
		c.logger.Log(logger.Error, "deploy: recovery failed", "error: "+err.Error())
		if err == types.ErrS3NotConfigured {
			return nil, fuego.BadRequestError{
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

	msg := "Recovery completed"
	status := "success"
	if len(result.Failed) > 0 && len(result.Recovered) == 0 {
		status = "failed"
		msg = "All recoveries failed"
	} else if len(result.Failed) > 0 {
		status = "partial"
		msg = "Some applications failed to recover"
	}

	c.logger.Log(logger.Info, "deploy: recovery process finished",
		fmt.Sprintf("recovered: %d, skipped: %d, failed: %d", len(result.Recovered), len(result.Skipped), len(result.Failed)))

	return &types.RecoverResponse{
		Status:  status,
		Message: msg,
		Data:    *result,
	}, nil
}
