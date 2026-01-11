package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *HealthCheckController) DeleteHealthCheck(f fuego.ContextNoBody) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{Status: http.StatusUnauthorized}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		return nil, fuego.HTTPError{Status: http.StatusBadRequest}
	}

	q := r.URL.Query()
	applicationID := q.Get("application_id")
	if applicationID == "" {
		return nil, fuego.HTTPError{Status: http.StatusBadRequest, Err: types.ErrInvalidApplicationID}
	}

	if err := c.service.DeleteHealthCheck(applicationID, orgID); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		statusCode, mappedErr := mapHealthCheckError(err)
		return &shared_types.Response{
			Status: "error",
			Error:  mappedErr.Error(),
		}, fuego.HTTPError{Status: statusCode}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Health check deleted successfully",
		Data:    nil,
	}, nil
}
