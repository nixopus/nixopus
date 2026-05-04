package controller

import (
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

func (c *MachineController) TriggerBackup(f fuego.ContextNoBody) (*types.TriggerBackupResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logMachineDebug("TriggerBackup", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("TriggerBackup", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	serverID := parseServerID(r)
	resp, err := c.backupService.TriggerBackup(r.Context(), user.ID, orgID, serverID)
	if err != nil {
		data := fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID)
		if serverID != nil {
			data = fmt.Sprintf("%s server_id=%s", data, *serverID)
		}
		c.logger.Log(logger.Error, fmt.Sprintf("machine: TriggerBackup: %v", err), data)
		return nil, mapBackupError(err)
	}

	return resp, nil
}

func (c *MachineController) ListBackups(f fuego.ContextNoBody) (*types.BackupListResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logMachineDebug("ListBackups", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("ListBackups", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	var params types.BackupListParams
	if v := f.QueryParam("page"); v != "" {
		if page, err := strconv.Atoi(v); err == nil && page > 0 {
			params.Page = page
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
	if v := f.QueryParam("server_id"); v != "" {
		params.ServerID = strings.TrimSpace(v)
	}

	response, err := c.backupService.ListBackups(r.Context(), orgID, params)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("machine: ListBackups: %v", err), fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.HTTPError{Detail: "failed to list backups", Status: http.StatusInternalServerError}
	}

	return response, nil
}

func (c *MachineController) GetBackupSchedule(f fuego.ContextNoBody) (*types.BackupScheduleResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logMachineDebug("GetBackupSchedule", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("GetBackupSchedule", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	serverID := parseServerID(r)
	response, err := c.backupService.GetBackupSchedule(r.Context(), orgID, serverID)
	if err != nil {
		data := fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID)
		if serverID != nil {
			data = fmt.Sprintf("%s server_id=%s", data, *serverID)
		}
		c.logger.Log(logger.Error, fmt.Sprintf("machine: GetBackupSchedule: %v", err), data)
		return nil, fuego.HTTPError{Detail: "failed to get backup schedule", Status: http.StatusInternalServerError}
	}

	return response, nil
}

func (c *MachineController) UpdateBackupSchedule(f fuego.ContextWithBody[types.BackupScheduleData]) (*types.BackupScheduleResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logMachineDebug("UpdateBackupSchedule", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("UpdateBackupSchedule", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logMachineDebug("UpdateBackupSchedule", "invalid body", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid request body"}
	}

	serverID := parseServerID(r)
	response, err := c.backupService.UpdateBackupSchedule(r.Context(), orgID, serverID, body)
	if err != nil {
		data := fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID)
		if serverID != nil {
			data = fmt.Sprintf("%s server_id=%s", data, *serverID)
		}
		c.logger.Log(logger.Error, fmt.Sprintf("machine: UpdateBackupSchedule: %v", err), data)
		return nil, fuego.BadRequestError{Detail: err.Error()}
	}

	return response, nil
}

func mapBackupError(err error) error {
	switch {
	case errors.Is(err, types.ErrMachineNotProvisioned):
		return fuego.NotFoundError{Detail: err.Error()}
	case errors.Is(err, types.ErrBackupAlreadyRunning):
		return fuego.HTTPError{Detail: "a backup is already in progress", Status: http.StatusConflict}
	case errors.Is(err, types.ErrS3NotConfigured):
		return fuego.BadRequestError{Detail: err.Error()}
	default:
		return fuego.HTTPError{Detail: "backup operation failed", Status: http.StatusInternalServerError}
	}
}
