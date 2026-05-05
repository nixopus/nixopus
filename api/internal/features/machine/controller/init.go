package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/service"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	sharedtypes "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"

	ff_service "github.com/nixopus/nixopus/api/internal/features/feature-flags/service"
	ff_storage "github.com/nixopus/nixopus/api/internal/features/feature-flags/storage"
)

type MachineController struct {
	store               *shared_storage.Store
	service             *service.MachineService
	listService         *service.ListService
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
	bs.Logger = &l
	backupStore := machine_storage.NewBackupStorage(store.DB, ctx)
	backupStore.Logger = &l
	regStore := machine_storage.NewRegistrationStorage(store.DB, ctx)
	regStore.Logger = &l
	listStore := machine_storage.NewListStorage(store.DB, ctx)
	listStore.Logger = &l
	ffStorage := ff_storage.NewFeatureFlagStorage(store.DB, ctx)
	ffService := ff_service.NewFeatureFlagService(ffStorage, l, ctx)
	regService := service.NewRegistrationService(regStore, ffService, nil, l, ctx)
	return &MachineController{
		store:               store,
		service:             service.NewMachineService(store, ctx, l, regStore),
		listService:         service.NewListService(listStore, l, ctx),
		billingService:      service.NewBillingService(bs),
		lifecycleService:    service.NewLifecycleService(bs, queue.ExecuteMachineLifecycle),
		backupService:       service.NewBackupService(bs, backupStore, store.DB, config.AppConfig.S3),
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
		c.logMachineDebug("GetSystemStats", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("GetSystemStats", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	ctx := r.Context()
	if sid := parseServerID(r); sid != nil {
		ctx = context.WithValue(ctx, sharedtypes.ServerIDKey, sid.String())
	}

	response, err := c.service.GetSystemStats(ctx, orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("machine: GetSystemStats: %v", err), fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
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
		c.logMachineDebug("ExecCommand", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("ExecCommand", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logMachineDebug("ExecCommand", "invalid body", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid request body"}
	}

	if body.Command == "" {
		c.logMachineDebug("ExecCommand", "command required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "command is required"}
	}

	ctx := r.Context()
	if sid := parseServerID(r); sid != nil {
		ctx = context.WithValue(ctx, sharedtypes.ServerIDKey, sid.String())
	}

	response, err := c.service.ExecCommand(ctx, orgID, body.Command)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("machine: ExecCommand: %v", err), fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return response, nil
}
