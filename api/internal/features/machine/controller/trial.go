package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"

	"github.com/nixopus/nixopus/api/internal/cache"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_service "github.com/nixopus/nixopus/api/internal/features/machine/service"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/features/machine/validation"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// TrailController handles HTTP requests for trial machine provisioning.
type TrailController struct {
	validator *validation.Validator
	service   *machine_service.TrailService
	ctx       context.Context
	logger    logger.Logger
	cache     *cache.Cache
}

// NewTrailController creates a new TrailController instance.
func NewTrailController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	c *cache.Cache,
) *TrailController {
	trailStorage := machine_storage.NewTrailStorage(store.DB, ctx)

	return &TrailController{
		validator: validation.NewValidator(),
		service:   machine_service.NewTrailService(store, ctx, l, trailStorage),
		ctx:       ctx,
		logger:    l,
		cache:     c,
	}
}

func mapTrialErrorToStatus(err error) int {
	switch {
	case errors.Is(err, machine_types.ErrImageNotAllowed):
		return http.StatusBadRequest
	case errors.Is(err, machine_types.ErrActiveProvisionExists):
		return http.StatusConflict
	case errors.Is(err, machine_types.ErrSystemAtCapacity):
		return http.StatusServiceUnavailable
	case errors.Is(err, machine_types.ErrProvisionNotFound):
		return http.StatusNotFound
	case errors.Is(err, machine_types.ErrInvalidSessionID):
		return http.StatusBadRequest
	case errors.Is(err, machine_types.ErrOrganizationRequired):
		return http.StatusForbidden
	case errors.Is(err, machine_types.ErrInvalidOrganizationID):
		return http.StatusBadRequest
	case errors.Is(err, machine_types.ErrFailedToEnqueueTask):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ProvisionTrail handles POST /api/v1/machines/trial/provision
func (c *TrailController) ProvisionTrail(f fuego.ContextWithBody[machine_types.ProvisionRequest]) (*machine_types.ProvisionTrailResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := r.Header.Get("X-Organization-Id")
	if orgID == "" {
		return nil, fuego.ForbiddenError{Detail: machine_types.ErrOrganizationRequired.Error(), Err: machine_types.ErrOrganizationRequired}
	}

	if _, err := uuid.Parse(orgID); err != nil {
		return nil, fuego.BadRequestError{Detail: machine_types.ErrInvalidOrganizationID.Error(), Err: machine_types.ErrInvalidOrganizationID}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), user.ID.String())
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if err := c.validator.ValidateRequest(&body); err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	result, err := c.service.ProvisionTrail(user.ID.String(), orgID, body)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), user.ID.String())
		status := mapTrialErrorToStatus(err)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: status,
		}
	}

	if c.cache != nil {
		_ = c.cache.InvalidateUserByID(c.ctx, user.ID.String())
	}

	return &machine_types.ProvisionTrailResponse{
		Status:  "success",
		Message: "Trail provisioning started",
		Data:    result,
	}, nil
}

// GetStatus handles GET /api/v1/machines/trial/status/{sessionId}
func (c *TrailController) GetStatus(f fuego.ContextNoBody) (*machine_types.TrailStatusEnvelopeResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	sessionID := f.PathParam("sessionId")
	if sessionID == "" {
		return nil, fuego.BadRequestError{Detail: machine_types.ErrInvalidSessionID.Error(), Err: machine_types.ErrInvalidSessionID}
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, fuego.BadRequestError{Detail: machine_types.ErrInvalidSessionID.Error(), Err: machine_types.ErrInvalidSessionID}
	}

	result, err := c.service.GetStatus(user.ID.String(), sessionID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), user.ID.String())
		status := mapTrialErrorToStatus(err)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: status,
		}
	}

	return &machine_types.TrailStatusEnvelopeResponse{
		Status:  "success",
		Message: "Status retrieved successfully",
		Data:    result,
	}, nil
}

// UpgradeResources handles POST /trail/upgrade-resources (internal route)
func (c *TrailController) UpgradeResources(f fuego.ContextWithBody[machine_types.UpgradeResourcesRequest]) (*machine_types.UpgradeResourcesResponse, error) {
	r := f.Request()

	secret := r.Header.Get("X-Internal-Secret")
	if secret == "" || secret != config.AppConfig.BetterAuth.Secret {
		return nil, fuego.UnauthorizedError{Detail: "unauthorized", Err: errors.New("unauthorized")}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if body.UserID == "" || body.OrgID == "" {
		return nil, fuego.BadRequestError{Detail: "user_id and org_id are required", Err: errors.New("user_id and org_id are required")}
	}

	if body.VcpuCount <= 0 || body.MemoryMB <= 0 {
		return nil, fuego.BadRequestError{Detail: "vcpu_count and memory_mb must be positive", Err: errors.New("vcpu_count and memory_mb must be positive")}
	}

	if err := c.service.UpgradeResources(body.UserID, body.OrgID, body.VcpuCount, body.MemoryMB); err != nil {
		c.logger.Log(logger.Error, err.Error(), body.UserID)
		status := mapTrialErrorToStatus(err)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: status,
		}
	}

	return &machine_types.UpgradeResourcesResponse{
		Status:  "success",
		Message: "Resource upgrade enqueued",
	}, nil
}
