package controller

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *MachineController) ListMachines(f fuego.ContextNoBody) (*types.ListMachinesResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("ListMachines", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("ListMachines", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	params := types.MachineListParams{}

	if v := f.QueryParam("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			params.Page = p
		}
	}
	if v := f.QueryParam("page_size"); v != "" {
		if ps, err := strconv.Atoi(v); err == nil && ps > 0 {
			params.PageSize = ps
		}
	}
	if v := f.QueryParam("search"); v != "" {
		params.Search = strings.TrimSpace(v)
	}
	if v := f.QueryParam("sort_by"); v != "" {
		params.SortBy = strings.ToLower(strings.TrimSpace(v))
	}
	if v := f.QueryParam("sort_order"); v != "" {
		params.SortOrder = strings.ToLower(strings.TrimSpace(v))
	}
	if v := f.QueryParam("status"); v != "" {
		params.Status = strings.TrimSpace(v)
	}
	if v := f.QueryParam("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			params.IsActive = &b
		}
	}

	response, err := c.listService.ListMachines(orgID, params)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("machine: ListMachines: %v", err), fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return response, nil
}

func (c *MachineController) CheckSSHStatus(f fuego.ContextNoBody) (*types.SSHConnectionStatusResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("CheckSSHStatus", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("CheckSSHStatus", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	response, err := c.listService.CheckSSHConnection(orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("machine: CheckSSHStatus: %v", err), fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return response, nil
}

func (c *MachineController) SetDefaultMachine(f fuego.ContextNoBody) (*types.SetDefaultMachineResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("SetDefaultMachine", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("SetDefaultMachine", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	machineID, err := uuid.Parse(f.PathParam("id"))
	if err != nil {
		c.logMachineDebug("SetDefaultMachine", "invalid machine ID", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid machine ID"}
	}

	key, err := c.listService.SetDefaultMachine(orgID, machineID)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrMachineNotFound), errors.Is(err, sql.ErrNoRows):
			return nil, fuego.NotFoundError{Detail: "machine not found"}
		case errors.Is(err, types.ErrMachineInactive):
			return nil, fuego.BadRequestError{Detail: err.Error()}
		default:
			c.logger.Log(logger.Error, fmt.Sprintf("machine: SetDefaultMachine: %v", err), fmt.Sprintf("org_id=%s user_id=%s machine_id=%s", orgID, user.ID, machineID))
			return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
		}
	}

	return &types.SetDefaultMachineResponse{
		Status:  "success",
		Message: "Server set as default successfully",
		Data:    *key,
	}, nil
}
