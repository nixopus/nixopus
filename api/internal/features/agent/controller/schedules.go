package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/scheduler"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type ScheduleResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (c *AgentController) ListSchedules(f fuego.ContextNoBody) (*ScheduleResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	schedules, err := c.service.GetScheduleStore().ListSchedulesForUser(r.Context(), user.ID.String(), orgID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &ScheduleResponse{
		Status:  "success",
		Message: "Schedules fetched successfully",
		Data:    schedules,
	}, nil
}

func (c *AgentController) GetSchedule(f fuego.ContextNoBody) (*ScheduleResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	scheduleID, err := uuid.Parse(f.PathParam("scheduleId"))
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid schedule ID"}
	}

	schedule, err := c.service.GetScheduleStore().GetScheduleForUser(r.Context(), scheduleID, user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.NotFoundError{Detail: "schedule not found", Err: err}
	}

	return &ScheduleResponse{
		Status:  "success",
		Message: "Schedule fetched successfully",
		Data:    schedule,
	}, nil
}

func (c *AgentController) GetScheduleRuns(f fuego.ContextNoBody) (*ScheduleResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	scheduleID, err := uuid.Parse(f.PathParam("scheduleId"))
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid schedule ID"}
	}

	if _, err := c.service.GetScheduleStore().GetScheduleForUser(r.Context(), scheduleID, user.ID.String()); err != nil {
		return nil, fuego.NotFoundError{Detail: "schedule not found", Err: err}
	}

	runs, err := c.service.GetScheduleStore().GetRunsForSchedule(r.Context(), scheduleID, 50)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &ScheduleResponse{
		Status:  "success",
		Message: "Schedule runs fetched successfully",
		Data:    runs,
	}, nil
}

type UpdateScheduleStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active paused" description:"Target status: active or paused"`
}

func (c *AgentController) UpdateScheduleStatus(f fuego.ContextWithBody[UpdateScheduleStatusRequest]) (*ScheduleResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	scheduleID, err := uuid.Parse(f.PathParam("scheduleId"))
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid schedule ID"}
	}

	body, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	sched, err := c.service.GetScheduleStore().GetScheduleForUser(r.Context(), scheduleID, user.ID.String())
	if err != nil {
		return nil, fuego.NotFoundError{Detail: "schedule not found", Err: err}
	}

	status := scheduler.ScheduleStatus(body.Status)
	if err := c.service.GetScheduleStore().UpdateScheduleStatus(r.Context(), sched.ID, status); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &ScheduleResponse{
		Status:  "success",
		Message: "Schedule status updated",
		Data:    map[string]string{"id": sched.ID.String(), "status": body.Status},
	}, nil
}

func (c *AgentController) DeleteSchedule(f fuego.ContextNoBody) (*ScheduleResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	scheduleID, err := uuid.Parse(f.PathParam("scheduleId"))
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid schedule ID"}
	}

	if err := c.service.GetScheduleStore().SoftDeleteSchedule(r.Context(), scheduleID, user.ID.String()); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &ScheduleResponse{
		Status:  "success",
		Message: "Schedule deleted",
	}, nil
}
