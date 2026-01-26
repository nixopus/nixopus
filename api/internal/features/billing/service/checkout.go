package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/billing/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stripe/stripe-go/v76"
	billingportal_session "github.com/stripe/stripe-go/v76/billingportal/session"
	checkout_session "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
)

// CreateCheckoutSession creates a Stripe checkout session for upgrading to a paid plan
// For Indian export compliance: includes customer name, description, and billing address collection
func (s *BillingService) CreateCheckoutSession(organizationID uuid.UUID, userName string, userEmail string, successURL string, cancelURL string) (*shared_types.CheckoutSessionResponse, error) {
	if !s.IsConfigured() {
		return nil, types.ErrStripeNotConfigured
	}

	if s.config.PriceID == "" {
		return nil, fmt.Errorf("stripe price ID is not configured")
	}

	// Get or create billing account
	billingAccount, err := s.getOrCreateBillingAccount(organizationID)
	if err != nil {
		return nil, err
	}

	// Get or create Stripe customer with name (required for Indian export compliance)
	customerID, err := s.getOrCreateStripeCustomer(billingAccount, userName, userEmail, organizationID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get/create Stripe customer", err.Error())
		return nil, err
	}

	// Create checkout session with description (required for Indian export compliance)
	// Stripe Checkout will collect billing address automatically
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(s.config.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		// Collect billing address for Indian export compliance (required for international payments)
		BillingAddressCollection: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionRequired)),
		// Subscription description for Indian export compliance
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Description: stripe.String("Nixopus Pro Plan Subscription - Cloud deployment platform service"),
		},
		Metadata: map[string]string{
			"organization_id":    organizationID.String(),
			"billing_account_id": billingAccount.ID.String(),
		},
	}

	sess, err := checkout_session.New(params)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to create checkout session", err.Error())
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return &shared_types.CheckoutSessionResponse{
		CheckoutURL: sess.URL,
		SessionID:   sess.ID,
	}, nil
}

// CreateBillingPortalSession creates a Stripe billing portal session for managing subscription
func (s *BillingService) CreateBillingPortalSession(organizationID uuid.UUID, returnURL string) (*shared_types.BillingPortalResponse, error) {
	if !s.IsConfigured() {
		return nil, types.ErrStripeNotConfigured
	}

	// Get billing account
	billingAccount, err := s.storage.GetBillingAccountByOrgID(organizationID)
	if err != nil {
		return nil, err
	}

	if billingAccount.StripeCustomerID == nil {
		return nil, types.ErrCustomerNotFound
	}

	// Create portal session
	params := &stripe.BillingPortalSessionParams{
		Customer:  billingAccount.StripeCustomerID,
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := billingportal_session.New(params)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to create portal session", err.Error())
		return nil, fmt.Errorf("failed to create portal session: %w", err)
	}

	return &shared_types.BillingPortalResponse{
		PortalURL: sess.URL,
	}, nil
}

// getOrCreateStripeCustomer gets or creates a Stripe customer for the billing account
// For Indian export compliance: includes customer name (required for international payments)
func (s *BillingService) getOrCreateStripeCustomer(billingAccount *shared_types.BillingAccount, name string, email string, organizationID uuid.UUID) (string, error) {
	if billingAccount.StripeCustomerID != nil {
		// Update existing customer with name if not set (for Indian export compliance)
		if name != "" {
			custParams := &stripe.CustomerParams{
				Name: stripe.String(name),
			}
			_, err := customer.Update(*billingAccount.StripeCustomerID, custParams)
			if err != nil {
				s.logger.Log(logger.Error, "Failed to update customer name", err.Error())
				// Don't fail, continue with existing customer
			}
		}
		return *billingAccount.StripeCustomerID, nil
	}

	// Create new Stripe customer with name (required for Indian export compliance)
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"organization_id":    organizationID.String(),
			"billing_account_id": billingAccount.ID.String(),
		},
	}

	// Add name if provided (required for Indian export compliance)
	if name != "" {
		params.Name = stripe.String(name)
	}

	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Update billing account with customer ID
	billingAccount.StripeCustomerID = &cust.ID
	if err := s.storage.UpdateBillingAccount(billingAccount); err != nil {
		s.logger.Log(logger.Error, "Failed to update billing account with customer ID", err.Error())
		// Don't fail the operation, customer was created successfully
	}

	return cust.ID, nil
}

// GetInvoices retrieves invoices for an organization
func (s *BillingService) GetInvoices(organizationID uuid.UUID, limit int) ([]shared_types.Invoice, error) {
	billingAccount, err := s.storage.GetBillingAccountByOrgID(organizationID)
	if err != nil {
		if err == types.ErrBillingAccountNotFound {
			return []shared_types.Invoice{}, nil
		}
		return nil, err
	}

	return s.storage.GetInvoicesByBillingAccountID(billingAccount.ID, limit)
}
