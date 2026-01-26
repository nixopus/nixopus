package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/billing/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stripe/stripe-go/v76"
)

type BillingService struct {
	storage storage.BillingRepository
	Ctx     context.Context
	logger  logger.Logger
	config  types.StripeConfig
}

func NewBillingService(
	ctx context.Context,
	l logger.Logger,
	billingStorage storage.BillingRepository,
	config types.StripeConfig,
) *BillingService {
	// Initialize Stripe with the secret key
	if config.SecretKey != "" {
		stripe.Key = config.SecretKey
	}

	return &BillingService{
		storage: billingStorage,
		Ctx:     ctx,
		logger:  l,
		config:  config,
	}
}

// IsConfigured returns true if Stripe is properly configured
func (s *BillingService) IsConfigured() bool {
	return s.config.SecretKey != "" && s.config.PriceID != ""
}
