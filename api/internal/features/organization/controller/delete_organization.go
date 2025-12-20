package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// TODO: deleting the organization should only happen if there is no deployments or etc .... else do not allow it.
func (c *OrganizationsController) DeleteOrganization(f fuego.ContextWithBody[types.DeleteOrganizationRequest]) (*shared_types.Response, error) {
	organization, err := f.Body()
	c.logger.Log(logger.Info, "Deleting organization", organization.ID)

	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	if err := c.validator.ValidateRequest(&organization); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	loggedInUser := utils.GetUser(w, r)
	if loggedInUser == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID, err := uuid.Parse(organization.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	count, err := c.service.GetOrganizationCount(loggedInUser.ID.String())
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	if count <= 1 {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	if err := c.service.DeleteOrganization(organizationID); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// c.Notify(notification.NotificationPayloadTypeDeleteOrganization, loggedInUser, r)

	return &shared_types.Response{
		Status:  "success",
		Message: "Organization deleted successfully",
		Data:    nil,
	}, nil
}
