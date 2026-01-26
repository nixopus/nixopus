package controller

import (
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/billing/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// CreateCheckoutSession creates a Stripe checkout session for subscription upgrade
func (c *BillingController) CreateCheckoutSession(f fuego.ContextWithBody[types.CreateCheckoutRequest]) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "Failed to parse request body", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if err := c.validator.ValidateRequest(&body); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	organizationID := utils.GetOrganizationID(r)

	// Get user name (required for Indian export compliance)
	userName := user.Name
	if userName == "" {
		// Fallback to email prefix if name is not available
		if atIndex := strings.Index(user.Email, "@"); atIndex > 0 {
			userName = user.Email[:atIndex]
		} else {
			userName = user.Email
		}
	}

	result, err := c.service.CreateCheckoutSession(organizationID, userName, user.Email, body.SuccessURL, body.CancelURL)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to create checkout session", err.Error())

		if err == types.ErrStripeNotConfigured {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusServiceUnavailable,
			}
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Checkout session created successfully",
		Data:    result,
	}, nil
}

// CreateBillingPortal creates a Stripe billing portal session for managing subscription
func (c *BillingController) CreateBillingPortal(f fuego.ContextNoBody) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(r)

	// Get return URL from query params, default to referer
	returnURL := r.URL.Query().Get("return_url")
	if returnURL == "" {
		returnURL = r.Referer()
	}

	result, err := c.service.CreateBillingPortalSession(organizationID, returnURL)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to create billing portal session", err.Error())

		if err == types.ErrStripeNotConfigured {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusServiceUnavailable,
			}
		}

		if err == types.ErrCustomerNotFound {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Billing portal session created successfully",
		Data:    result,
	}, nil
}
