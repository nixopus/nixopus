package controller

import (
	"context"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/service"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/utils"

	ff_service "github.com/nixopus/nixopus/api/internal/features/feature-flags/service"
	ff_storage "github.com/nixopus/nixopus/api/internal/features/feature-flags/storage"
)

type MachineController struct {
	store               *shared_storage.Store
	service             *service.MachineService
	billingService      *service.BillingService
	lifecycleService    *service.LifecycleService
	backupService       *service.BackupService
	metricsService      *service.MetricsService
	registrationService *service.RegistrationService
	ctx                 context.Context
	logger              logger.Logger
}

func NewMachineController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	ts *machine_storage.TimescaleStore,
) *MachineController {
	bs := machine_storage.NewBillingStorage(store.DB, ctx)
	backupStore := machine_storage.NewBackupStorage(store.DB, ctx)
	regStore := machine_storage.NewRegistrationStorage(store.DB, ctx)
	ffStorage := ff_storage.NewFeatureFlagStorage(store.DB, ctx)
	ffService := ff_service.NewFeatureFlagService(ffStorage, l, ctx)
	regService := service.NewRegistrationService(regStore, ffService, nil, l, ctx)
	return &MachineController{
		store:               store,
		service:             service.NewMachineService(store, ctx, l),
		billingService:      service.NewBillingService(bs),
		lifecycleService:    service.NewLifecycleService(bs, queue.ExecuteMachineLifecycle),
		backupService:       service.NewBackupService(bs, backupStore, store.DB),
		metricsService:      service.NewMetricsService(ts, store.DB),
		registrationService: regService,
		ctx:                 ctx,
		logger:              l,
	}
}

func (c *MachineController) GetSystemStats(f fuego.ContextNoBody) (*types.SystemStatsResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "Organization ID not found in context", "")
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	response, err := c.service.GetSystemStats(orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return response, nil
}

func (c *MachineController) ExecCommand(f fuego.ContextWithBody[types.HostExecRequest]) (*types.HostExecResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "Organization ID not found in context", "")
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid request body"}
	}

	if body.Command == "" {
		return nil, fuego.BadRequestError{Detail: "command is required"}
	}

	response, err := c.service.ExecCommand(orgID, body.Command)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return response, nil
}
