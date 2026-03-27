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

func (c *MachineController) TriggerBackup(f fuego.ContextNoBody) (*types.TriggerBackupResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	response, err := c.backupService.TriggerBackup(r.Context(), user.ID, orgID)
	if err != nil {
		return nil, mapBackupError(err)
	}

	return response, nil
}

func (c *MachineController) ListBackups(f fuego.ContextNoBody) (*types.BackupListResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	response, err := c.backupService.ListBackups(r.Context(), orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{Detail: "failed to list backups", Status: http.StatusInternalServerError}
	}

	return response, nil
}

func (c *MachineController) GetBackupSchedule(f fuego.ContextNoBody) (*types.BackupScheduleResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	response, err := c.backupService.GetBackupSchedule(r.Context(), orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{Detail: "failed to get backup schedule", Status: http.StatusInternalServerError}
	}

	return response, nil
}

func (c *MachineController) UpdateBackupSchedule(f fuego.ContextWithBody[types.BackupScheduleData]) (*types.BackupScheduleResponse, error) {
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
		return nil, fuego.BadRequestError{Detail: "invalid request body"}
	}

	response, err := c.backupService.UpdateBackupSchedule(r.Context(), orgID, body)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
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
	default:
		return fuego.HTTPError{Detail: "backup operation failed", Status: http.StatusInternalServerError}
	}
}
