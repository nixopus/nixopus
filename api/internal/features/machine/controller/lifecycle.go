package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	sharedtypes "github.com/nixopus/nixopus/api/internal/types"
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
		c.logMachineDebug("GetMachineStatus", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("GetMachineStatus", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	serverID := parseServerID(r)
	if serverID == nil {
		c.logMachineDebug("GetMachineStatus", "server_id required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "server_id is required"}
	}

	ctx := context.WithValue(r.Context(), sharedtypes.ServerIDKey, serverID.String())
	response, err := c.service.GetMachineStatus(ctx, orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("machine: GetMachineStatus: %v", err), fmt.Sprintf("org_id=%s user_id=%s server_id=%s", orgID, user.ID, serverID))
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}
	return response, nil
}

func (c *MachineController) RestartMachine(f fuego.ContextNoBody) (*types.MachineActionResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logMachineDebug("RestartMachine", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("RestartMachine", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	serverID := parseServerID(r)
	if serverID == nil {
		c.logMachineDebug("RestartMachine", "server_id required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "server_id is required"}
	}

	userOwned, _ := c.billingService.IsServerUserOwned(orgID, *serverID)
	if userOwned {
		ctx := context.WithValue(r.Context(), sharedtypes.ServerIDKey, serverID.String())
		response, err := c.service.RestartMachine(ctx, orgID)
		if err != nil {
			c.logger.Log(logger.Error, fmt.Sprintf("machine: RestartMachine (user-owned): %v", err), fmt.Sprintf("org_id=%s user_id=%s server_id=%s", orgID, user.ID, serverID))
			return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
		}
		return response, nil
	}

	response, err := c.lifecycleService.Restart(r.Context(), orgID, serverID)
	if err != nil {
		return nil, mapLifecycleError(c.logger, err, orgID, user.ID, serverID, "restart")
	}

	return response, nil
}

func (c *MachineController) PauseMachine(f fuego.ContextNoBody) (*types.MachineActionResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logMachineDebug("PauseMachine", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("PauseMachine", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	serverID := parseServerID(r)
	if serverID == nil {
		c.logMachineDebug("PauseMachine", "server_id required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "server_id is required"}
	}

	userOwned, _ := c.billingService.IsServerUserOwned(orgID, *serverID)
	if userOwned {
		response, err := c.service.PauseMachine(r.Context(), orgID, *serverID)
		if err != nil {
			c.logger.Log(logger.Error, fmt.Sprintf("machine: PauseMachine (user-owned): %v", err), fmt.Sprintf("org_id=%s user_id=%s server_id=%s", orgID, user.ID, serverID))
			return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
		}
		return response, nil
	}

	response, err := c.lifecycleService.Pause(r.Context(), orgID, serverID)
	if err != nil {
		return nil, mapLifecycleError(c.logger, err, orgID, user.ID, serverID, "pause")
	}

	return response, nil
}

func (c *MachineController) ResumeMachine(f fuego.ContextNoBody) (*types.MachineActionResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logMachineDebug("ResumeMachine", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("ResumeMachine", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	serverID := parseServerID(r)
	if serverID == nil {
		c.logMachineDebug("ResumeMachine", "server_id required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "server_id is required"}
	}

	userOwned, _ := c.billingService.IsServerUserOwned(orgID, *serverID)
	if userOwned {
		response, err := c.service.ResumeMachine(r.Context(), orgID, *serverID)
		if err != nil {
			c.logger.Log(logger.Error, fmt.Sprintf("machine: ResumeMachine (user-owned): %v", err), fmt.Sprintf("org_id=%s user_id=%s server_id=%s", orgID, user.ID, serverID))
			return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
		}
		return response, nil
	}

	response, err := c.lifecycleService.Resume(r.Context(), orgID, serverID)
	if err != nil {
		return nil, mapLifecycleError(c.logger, err, orgID, user.ID, serverID, "resume")
	}

	return response, nil
}

func mapLifecycleError(l logger.Logger, err error, orgID uuid.UUID, userID uuid.UUID, serverID *uuid.UUID, action string) error {
	data := fmt.Sprintf("org_id=%s user_id=%s", orgID, userID)
	if serverID != nil {
		data = fmt.Sprintf("%s server_id=%s", data, *serverID)
	}
	l.Log(logger.Error, fmt.Sprintf("machine: lifecycle %s: %v", action, err), data)

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
