package controller

import (
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func parseServerID(r *http.Request) *uuid.UUID {
	s := r.URL.Query().Get("server_id")
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func (c *MachineController) GetMachineStatus(f fuego.ContextNoBody) (*types.MachineStateResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	response, err := c.lifecycleService.GetStatus(r.Context(), orgID, parseServerID(r))
	if err != nil {
		return nil, mapLifecycleError(c.logger, err, orgID, "get status")
	}

	return response, nil
}

func (c *MachineController) RestartMachine(f fuego.ContextNoBody) (*types.MachineActionResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	response, err := c.lifecycleService.Restart(r.Context(), orgID, parseServerID(r))
	if err != nil {
		return nil, mapLifecycleError(c.logger, err, orgID, "restart")
	}

	return response, nil
}

func (c *MachineController) PauseMachine(f fuego.ContextNoBody) (*types.MachineActionResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	response, err := c.lifecycleService.Pause(r.Context(), orgID, parseServerID(r))
	if err != nil {
		return nil, mapLifecycleError(c.logger, err, orgID, "pause")
	}

	return response, nil
}

func (c *MachineController) ResumeMachine(f fuego.ContextNoBody) (*types.MachineActionResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	response, err := c.lifecycleService.Resume(r.Context(), orgID, parseServerID(r))
	if err != nil {
		return nil, mapLifecycleError(c.logger, err, orgID, "resume")
	}

	return response, nil
}

func mapLifecycleError(l logger.Logger, err error, orgID uuid.UUID, action string) error {
	l.Log(logger.Error, err.Error(), orgID.String())

	switch {
	case errors.Is(err, types.ErrMachineNotProvisioned):
		return fuego.NotFoundError{Detail: err.Error()}
	case errors.Is(err, types.ErrMachineOperationTimeout):
		return fuego.HTTPError{Detail: "machine operation timed out, please try again", Status: http.StatusGatewayTimeout}
	case errors.Is(err, types.ErrMachineOperationLocked):
		return fuego.HTTPError{Detail: "another operation is already in progress", Status: http.StatusConflict}
	case errors.Is(err, types.ErrMachineNotRunning):
		return fuego.HTTPError{Detail: "machine is not currently running", Status: http.StatusUnprocessableEntity}
	case errors.Is(err, types.ErrMachineAlreadyPaused):
		return fuego.HTTPError{Detail: "machine is already paused", Status: http.StatusUnprocessableEntity}
	case errors.Is(err, types.ErrMachineNotPaused):
		return fuego.HTTPError{Detail: "machine is not paused", Status: http.StatusUnprocessableEntity}
	default:
		return fuego.HTTPError{Detail: "machine " + action + " failed, please try again later", Status: http.StatusInternalServerError}
	}
}
