package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/execute/service"
	"github.com/raghavyuva/nixopus-api/internal/features/execute/types"
	"github.com/raghavyuva/nixopus-api/internal/features/execute/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// ExecuteController handles HTTP requests for command execution.
type ExecuteController struct {
	validator *validation.Validator
	service   *service.ExecuteService
	ctx       context.Context
	logger    logger.Logger
}

// NewExecuteController creates a new ExecuteController instance.
func NewExecuteController(
	ctx context.Context,
	l logger.Logger,
) *ExecuteController {
	return &ExecuteController{
		validator: validation.NewValidator(),
		service:   service.NewExecuteService(l),
		ctx:       ctx,
		logger:    l,
	}
}

// Execute handles POST /api/v1/execute
//
// This endpoint executes a whitelisted command via the abyss CLI.
// Only authenticated users can execute commands.
//
// Request Body:
//   - command: the command to execute (must be in whitelist)
//   - args: optional command arguments
//
// Returns:
//   - 200 OK: command executed successfully
//   - 400 Bad Request: invalid request or command not allowed
//   - 401 Unauthorized: authentication required
func (c *ExecuteController) Execute(f fuego.ContextWithBody[types.ExecuteRequest]) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    errors.New("authentication required"),
			Status: http.StatusUnauthorized,
		}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), user.ID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if err := c.validator.ValidateRequest(&body); err != nil {
		c.logger.Log(logger.Warning, err.Error(), user.ID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusForbidden,
		}
	}

	result, err := c.service.ExecuteCommand(user.ID.String(), body)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), user.ID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	status := http.StatusOK
	if result.ExitCode != 0 {
		status = http.StatusBadRequest
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Command executed",
		Data:    result,
	}, fuego.HTTPError{Status: status}
}
