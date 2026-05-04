package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *MachineController) CreateMachine(f fuego.ContextWithBody[types.CreateMachineRequest]) (*types.CreateMachineResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("CreateMachine", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("CreateMachine", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logMachineDebug("CreateMachine", "invalid body", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid request body"}
	}

	if body.Name == "" || body.Host == "" {
		c.logMachineDebug("CreateMachine", "name and host required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "name and host are required"}
	}

	response, err := c.registrationService.CreateMachine(orgID, user.ID, body)
	if err != nil {
		return nil, mapRegistrationError(c.logger, err, orgID, user.ID, nil)
	}

	return response, nil
}

func (c *MachineController) VerifyMachine(f fuego.ContextNoBody) (*types.VerifyMachineResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("VerifyMachine", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("VerifyMachine", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	machineID, err := uuid.Parse(f.PathParam("id"))
	if err != nil {
		c.logMachineDebug("VerifyMachine", "invalid machine ID", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid machine ID"}
	}

	response, err := c.registrationService.VerifyMachine(orgID, machineID)
	if err != nil {
		return nil, mapRegistrationError(c.logger, err, orgID, user.ID, &machineID)
	}

	return response, nil
}

func (c *MachineController) RenameMachine(f fuego.ContextWithBody[types.RenameMachineRequest]) (*types.RenameMachineResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("RenameMachine", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("RenameMachine", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	machineID, err := uuid.Parse(f.PathParam("id"))
	if err != nil {
		c.logMachineDebug("RenameMachine", "invalid machine ID", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid machine ID"}
	}

	body, err := f.Body()
	if err != nil {
		c.logMachineDebug("RenameMachine", "invalid body", fmt.Sprintf("org_id=%s user_id=%s machine_id=%s", orgID, user.ID, machineID))
		return nil, fuego.BadRequestError{Detail: "invalid request body"}
	}

	response, err := c.registrationService.RenameMachine(orgID, machineID, body.Name)
	if err != nil {
		return nil, mapRegistrationError(c.logger, err, orgID, user.ID, &machineID)
	}

	return response, nil
}

func (c *MachineController) DeleteMachine(f fuego.ContextNoBody) (*types.DeleteMachineResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("DeleteMachine", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("DeleteMachine", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	machineID, err := uuid.Parse(f.PathParam("id"))
	if err != nil {
		c.logMachineDebug("DeleteMachine", "invalid machine ID", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid machine ID"}
	}

	if err := c.registrationService.DeleteMachine(orgID, machineID); err != nil {
		return nil, mapRegistrationError(c.logger, err, orgID, user.ID, &machineID)
	}

	return &types.DeleteMachineResponse{Status: "deleted"}, nil
}

func (c *MachineController) GetSSHStatus(f fuego.ContextNoBody) (*types.SSHStatusResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("GetSSHStatus", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("GetSSHStatus", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	machineID, err := uuid.Parse(f.PathParam("id"))
	if err != nil {
		c.logMachineDebug("GetSSHStatus", "invalid machine ID", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid machine ID"}
	}

	response, err := c.registrationService.GetSSHStatus(orgID, machineID)
	if err != nil {
		return nil, mapRegistrationError(c.logger, err, orgID, user.ID, &machineID)
	}

	return response, nil
}

func mapRegistrationError(l logger.Logger, err error, orgID uuid.UUID, userID uuid.UUID, machineID *uuid.UUID) error {
	data := fmt.Sprintf("org_id=%s user_id=%s", orgID, userID)
	if machineID != nil {
		data = fmt.Sprintf("%s machine_id=%s", data, *machineID)
	}
	l.Log(logger.Error, fmt.Sprintf("machine: registration: %v", err), data)

	switch {
	case errors.Is(err, types.ErrFeatureDisabled):
		return fuego.ForbiddenError{Detail: err.Error()}
	case errors.Is(err, types.ErrMachineLimitReached):
		return fuego.ForbiddenError{Detail: err.Error()}
	case errors.Is(err, types.ErrDuplicateHost):
		return fuego.BadRequestError{Detail: err.Error()}
	case errors.Is(err, types.ErrNameRequired):
		return fuego.BadRequestError{Detail: err.Error()}
	case errors.Is(err, types.ErrNameTooLong):
		return fuego.BadRequestError{Detail: err.Error()}
	case errors.Is(err, types.ErrMachineHasApps):
		return fuego.ConflictError{Detail: err.Error()}
	case errors.Is(err, types.ErrInsufficientCredits):
		return fuego.HTTPError{Detail: err.Error(), Status: http.StatusPaymentRequired}
	default:
		return fuego.HTTPError{Detail: err.Error(), Status: http.StatusInternalServerError}
	}
}
