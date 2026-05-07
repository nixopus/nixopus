package controller

import (
	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/agent/service"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type CancelRequest struct {
	ThreadID string `json:"thread_id"`
}

type CancelResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (c *AgentController) CancelStream(f fuego.ContextWithBody[CancelRequest]) (*CancelResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if body.ThreadID == "" {
		return nil, fuego.BadRequestError{Detail: "thread_id is required"}
	}

	cancelled := service.CancelStream(body.ThreadID)
	if !cancelled {
		return &CancelResponse{
			Status:  "not_found",
			Message: "No active stream found for this thread",
		}, nil
	}

	return &CancelResponse{
		Status:  "cancelled",
		Message: "Stream cancelled successfully",
	}, nil
}
