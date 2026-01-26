package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetBillingStatus returns the billing status for the current organization
func (c *BillingController) GetBillingStatus(f fuego.ContextNoBody) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(r)
	status, err := c.service.GetBillingStatus(organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to get billing status", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Billing status retrieved successfully",
		Data:    status,
	}, nil
}

// CanDeploy checks if the organization can create a new deployment
func (c *BillingController) CanDeploy(f fuego.ContextNoBody) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID:= utils.GetOrganizationID(r)

	result, err := c.service.CanDeploy(organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to check deployment eligibility", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Deployment eligibility checked",
		Data:    result,
	}, nil
}
