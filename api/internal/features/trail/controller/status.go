package controller

import (
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/trail/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetStatus handles GET /api/v1/trail/status/{sessionId}
//
// This endpoint retrieves the current status of a trail provisioning request.
// Only the owner of the provision can access its status.
//
// Path Parameters:
//   - sessionId: the UUID of the provision session
//
// Returns:
//   - 200 OK: status retrieved successfully
//   - 400 Bad Request: invalid session ID format
//   - 401 Unauthorized: authentication required
//   - 404 Not Found: provision not found or not owned by user
func (c *TrailController) GetStatus(f fuego.ContextNoBody) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    errors.New("authentication required"),
			Status: http.StatusUnauthorized,
		}
	}

	sessionID := f.PathParam("sessionId")
	if sessionID == "" {
		return nil, fuego.HTTPError{
			Err:    types.ErrInvalidSessionID,
			Status: http.StatusBadRequest,
		}
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, fuego.HTTPError{
			Err:    types.ErrInvalidSessionID,
			Status: http.StatusBadRequest,
		}
	}

	result, err := c.service.GetStatus(user.ID.String(), sessionID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), user.ID.String())
		status := mapErrorToStatus(err)
		return nil, fuego.HTTPError{
			Err:    err,
			Status: status,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Status retrieved successfully",
		Data:    result,
	}, nil
}
