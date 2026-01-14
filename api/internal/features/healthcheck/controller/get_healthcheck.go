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

func (c *HealthCheckController) GetHealthCheck(f fuego.ContextNoBody) (*shared_types.Response, error) {
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

	healthCheck, err := c.service.GetHealthCheck(applicationID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}

	// Return success with null data if health check doesn't exist
	return &shared_types.Response{
		Status:  "success",
		Message: "Health check fetched successfully",
		Data:    healthCheck,
	}, nil
}
