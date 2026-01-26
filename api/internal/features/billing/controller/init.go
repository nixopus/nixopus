package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/billing/service"
	"github.com/raghavyuva/nixopus-api/internal/features/billing/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/billing/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type BillingController struct {
	store     *shared_storage.Store
	validator *validation.Validator
	service   *service.BillingService
	ctx       context.Context
	logger    logger.Logger
	config    types.StripeConfig
}

func NewBillingController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	config types.StripeConfig,
) *BillingController {
	billingStorage := &storage.BillingStorage{DB: store.DB, Ctx: ctx}
	return &BillingController{
		store:     store,
		validator: validation.NewValidator(),
		service:   service.NewBillingService(ctx, l, billingStorage, config),
		ctx:       ctx,
		logger:    l,
		config:    config,
	}
}
