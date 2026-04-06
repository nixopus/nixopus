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

func (c *MachineController) CreateMachine(f fuego.ContextWithBody[types.CreateMachineRequest]) (*types.CreateMachineResponse, error) {
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

	if body.Name == "" || body.Host == "" {
		return nil, fuego.BadRequestError{Detail: "name and host are required"}
	}

	response, err := c.registrationService.CreateMachine(orgID, user.ID, body)
	if err != nil {
		return nil, mapRegistrationError(c.logger, err, orgID)
	}

	return response, nil
}

func (c *MachineController) VerifyMachine(f fuego.ContextNoBody) (*types.VerifyMachineResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	machineID, err := uuid.Parse(f.PathParam("id"))
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid machine ID"}
	}

	if err := c.registrationService.VerifyMachine(orgID, machineID); err != nil {
		return nil, mapRegistrationError(c.logger, err, orgID)
	}

	return &types.VerifyMachineResponse{Status: "verification_queued"}, nil
}

func (c *MachineController) DeleteMachine(f fuego.ContextNoBody) (*types.DeleteMachineResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	machineID, err := uuid.Parse(f.PathParam("id"))
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid machine ID"}
	}

	if err := c.registrationService.DeleteMachine(orgID, machineID); err != nil {
		return nil, mapRegistrationError(c.logger, err, orgID)
	}

	return &types.DeleteMachineResponse{Status: "deleted"}, nil
}

func (c *MachineController) GetSSHStatus(f fuego.ContextNoBody) (*types.SSHStatusResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	machineID, err := uuid.Parse(f.PathParam("id"))
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid machine ID"}
	}

	response, err := c.registrationService.GetSSHStatus(orgID, machineID)
	if err != nil {
		return nil, mapRegistrationError(c.logger, err, orgID)
	}

	return response, nil
}

func mapRegistrationError(l logger.Logger, err error, orgID uuid.UUID) error {
	l.Log(logger.Error, err.Error(), orgID.String())

	switch {
	case errors.Is(err, types.ErrFeatureDisabled):
		return fuego.ForbiddenError{Detail: err.Error()}
	case errors.Is(err, types.ErrMachineLimitReached):
		return fuego.ForbiddenError{Detail: err.Error()}
	case errors.Is(err, types.ErrDuplicateHost):
		return fuego.BadRequestError{Detail: err.Error()}
	case errors.Is(err, types.ErrMachineHasApps):
		return fuego.ConflictError{Detail: err.Error()}
	case errors.Is(err, types.ErrInsufficientCredits):
		return fuego.HTTPError{Detail: err.Error(), Status: http.StatusPaymentRequired}
	default:
		return fuego.HTTPError{Detail: err.Error(), Status: http.StatusInternalServerError}
	}
}
