package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/billing/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CanDeploy checks if an organization can create a new deployment
// Returns true if:
// - The organization has an active subscription, OR
// - The organization has not exceeded its free deployment limit
func (s *BillingService) CanDeploy(organizationID uuid.UUID) (*shared_types.CanDeployResponse, error) {
	// Get or create billing account
	billingAccount, err := s.getOrCreateBillingAccount(organizationID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get/create billing account", err.Error())
		return nil, err
	}

	// Check for active subscription
	subscription, err := s.storage.GetActiveSubscription(billingAccount.ID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to check subscription", err.Error())
		return nil, err
	}

	// If there's an active subscription, allow unlimited deployments
	if subscription != nil {
		return &shared_types.CanDeployResponse{
			CanDeploy: true,
			Reason:    "Active subscription",
		}, nil
	}

	// Count current deployed applications
	deploymentCount, err := s.storage.CountDeployedApplications(organizationID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to count deployments", err.Error())
		return nil, err
	}

	// Check against free tier limit
	limit := billingAccount.FreeDeploymentsLimit
	if s.config.FreeDeploymentsLimit > 0 {
		limit = s.config.FreeDeploymentsLimit
	}

	if deploymentCount < limit {
		return &shared_types.CanDeployResponse{
			CanDeploy: true,
			Reason:    "Within free tier limit",
		}, nil
	}

	// Deployment limit reached
	return &shared_types.CanDeployResponse{
		CanDeploy: false,
		Reason:    types.ErrPaymentRequired.Error(),
	}, nil
}

// GetBillingStatus returns the current billing status for an organization
func (s *BillingService) GetBillingStatus(organizationID uuid.UUID) (*shared_types.BillingStatus, error) {
	// Get or create billing account
	billingAccount, err := s.getOrCreateBillingAccount(organizationID)
	if err != nil {
		return nil, err
	}

	// Get active subscription
	subscription, err := s.storage.GetActiveSubscription(billingAccount.ID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get subscription", err.Error())
		return nil, err
	}

	// Count deployments
	deploymentCount, err := s.storage.CountDeployedApplications(organizationID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to count deployments", err.Error())
		return nil, err
	}

	// Determine deployment limit
	limit := billingAccount.FreeDeploymentsLimit
	if s.config.FreeDeploymentsLimit > 0 {
		limit = s.config.FreeDeploymentsLimit
	}

	status := &shared_types.BillingStatus{
		HasActiveSubscription: subscription != nil,
		UsageStats: &shared_types.BillingUsageStats{
			DeploymentsUsed:  deploymentCount,
			DeploymentsLimit: limit,
			IsUnlimited:      subscription != nil,
		},
		BillingAccount: &shared_types.BillingAccountInfo{
			ID:                   billingAccount.ID.String(),
			FreeDeploymentsLimit: limit,
			HasStripeCustomer:    billingAccount.StripeCustomerID != nil,
		},
	}

	if subscription != nil {
		status.Subscription = &shared_types.SubscriptionInfo{
			ID:                 subscription.ID.String(),
			Status:             subscription.Status,
			CurrentPeriodStart: subscription.CurrentPeriodStart,
			CurrentPeriodEnd:   subscription.CurrentPeriodEnd,
			CancelAtPeriodEnd:  subscription.CancelAtPeriodEnd,
		}
	}

	return status, nil
}

// getOrCreateBillingAccount gets or creates a billing account for an organization
func (s *BillingService) getOrCreateBillingAccount(organizationID uuid.UUID) (*shared_types.BillingAccount, error) {
	account, err := s.storage.GetBillingAccountByOrgID(organizationID)
	if err == nil {
		return account, nil
	}

	// Account doesn't exist, create one
	if err == types.ErrBillingAccountNotFound {
		limit := 1
		if s.config.FreeDeploymentsLimit > 0 {
			limit = s.config.FreeDeploymentsLimit
		}

		newAccount := &shared_types.BillingAccount{
			OrganizationID:       organizationID,
			FreeDeploymentsLimit: limit,
		}

		if err := s.storage.CreateBillingAccount(newAccount); err != nil {
			return nil, err
		}

		return newAccount, nil
	}

	return nil, err
}
