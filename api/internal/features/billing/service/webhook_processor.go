package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/billing/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/billing/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/invoice"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"
)

// ProcessWebhook handles incoming Stripe webhook events
func (s *BillingService) ProcessWebhook(payload []byte, signature string) error {
	if !s.IsConfigured() || s.config.WebhookSecret == "" {
		return types.ErrStripeNotConfigured
	}

	// Verify webhook signature
	// Note: Stripe CLI uses API version 2020-08-27 by default, but stripe-go v76 expects 2023-10-16
	// We ignore the API version mismatch for local testing compatibility
	event, err := webhook.ConstructEventWithOptions(
		payload,
		signature,
		s.config.WebhookSecret,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to verify webhook signature", err.Error())
		return types.ErrInvalidWebhookSignature
	}

	// Check idempotency
	processed, err := s.storage.IsEventProcessed(event.ID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to check event idempotency", err.Error())
		return err
	}
	if processed {
		s.logger.Log(logger.Info, "Event already processed, skipping", event.ID)
		return nil
	}

	// Process the event
	if err := s.handleEvent(&event); err != nil {
		s.logger.Log(logger.Error, "Failed to process webhook event", err.Error())
		return err
	}

	// Record the event
	payloadJSON, _ := json.Marshal(event)
	paymentEvent := &shared_types.PaymentEvent{
		StripeEventID: event.ID,
		EventType:     string(event.Type),
		Payload:       payloadJSON,
	}
	if err := s.storage.RecordEvent(paymentEvent); err != nil {
		s.logger.Log(logger.Error, "Failed to record event", err.Error())
		// Don't fail the webhook, event was processed successfully
	}

	return nil
}

// handleEvent routes the event to the appropriate handler
// Following Stripe's snapshot event handler pattern: https://docs.stripe.com/webhooks
func (s *BillingService) handleEvent(event *stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutSessionCompleted(event)
	case "invoice.payment_succeeded":
		return s.handleInvoicePaymentSucceeded(event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(event)
	case "customer.subscription.created":
		// Handle subscription.created the same way as subscription.updated
		return s.handleSubscriptionUpdated(event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(event)
	default:
		s.logger.Log(logger.Info, "Unhandled webhook event type", string(event.Type))
		return nil
	}
}

// handleCheckoutSessionCompleted processes checkout.session.completed events
func (s *BillingService) handleCheckoutSessionCompleted(event *stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("failed to unmarshal checkout session: %w", err)
	}

	// Get organization ID from metadata
	orgIDStr, ok := session.Metadata["organization_id"]
	if !ok {
		s.logger.Log(logger.Error, "Missing organization_id in checkout session metadata", session.ID)
		return nil
	}

	organizationID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return fmt.Errorf("invalid organization_id: %w", err)
	}

	// Get billing account
	billingAccount, err := s.getOrCreateBillingAccount(organizationID)
	if err != nil {
		return err
	}

	// Update Stripe customer ID if not set
	if billingAccount.StripeCustomerID == nil && session.Customer != nil {
		billingAccount.StripeCustomerID = &session.Customer.ID
		if err := s.storage.UpdateBillingAccount(billingAccount); err != nil {
			s.logger.Log(logger.Error, "Failed to update billing account customer ID", err.Error())
		}
	}

	// For subscription checkouts, fetch and create the invoice immediately
	// Stripe creates an invoice when a subscription checkout is completed
	if session.Mode == stripe.CheckoutSessionModeSubscription && session.Subscription != nil {
		subscriptionID := session.Subscription.ID

		// Fetch the subscription to get the latest invoice
		sub, err := subscription.Get(subscriptionID, nil)
		if err != nil {
			s.logger.Log(logger.Error, "Failed to fetch subscription for invoice", err.Error())
		} else {
			// Use the invoice directly if it's already expanded, otherwise fetch it
			var stripeInvoice *stripe.Invoice
			if sub.LatestInvoice != nil {
				// LatestInvoice is already expanded as *stripe.Invoice
				stripeInvoice = sub.LatestInvoice
			} else {
				// Fetch the latest invoice for this subscription
				invoiceParams := &stripe.InvoiceListParams{
					Subscription: stripe.String(subscriptionID),
				}
				invoiceParams.Filters.AddFilter("limit", "", "1")
				invoices := invoice.List(invoiceParams)
				if invoices.Next() {
					stripeInvoice = invoices.Invoice()
				}
			}

			if stripeInvoice != nil {
				// Create invoice record in our database
				if err := s.createOrUpdateInvoiceFromStripe(stripeInvoice, billingAccount.ID); err != nil {
					s.logger.Log(logger.Error, "Failed to create invoice from checkout session", err.Error())
					// Don't fail the webhook, invoice.payment_succeeded will handle it
				} else {
					s.logger.Log(logger.Info, "Invoice created from checkout session", stripeInvoice.ID)
				}
			}
		}
	}

	// Subscription will be created/updated by subscription.updated event
	s.logger.Log(logger.Info, "Checkout session completed", session.ID)
	return nil
}

// handleInvoicePaymentSucceeded processes invoice.payment_succeeded events
func (s *BillingService) handleInvoicePaymentSucceeded(event *stripe.Event) error {
	var stripeInvoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &stripeInvoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Get billing account by customer ID
	billingAccount, err := s.getBillingAccountByCustomerID(stripeInvoice.Customer.ID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get billing account for invoice", err.Error())
		return nil // Don't fail, might be a different product
	}

	return s.createOrUpdateInvoiceFromStripe(&stripeInvoice, billingAccount.ID)
}

// createOrUpdateInvoiceFromStripe creates or updates an invoice record from a Stripe invoice
func (s *BillingService) createOrUpdateInvoiceFromStripe(stripeInvoice *stripe.Invoice, billingAccountID uuid.UUID) error {
	// Check if invoice already exists
	existingInvoice, _ := s.storage.GetInvoiceByStripeID(stripeInvoice.ID)

	// Map Stripe invoice status to our internal status
	var status shared_types.InvoiceStatus
	switch stripeInvoice.Status {
	case stripe.InvoiceStatusPaid:
		status = shared_types.InvoiceStatusPaid
	case stripe.InvoiceStatusOpen:
		status = shared_types.InvoiceStatusOpen
	case stripe.InvoiceStatusDraft:
		status = shared_types.InvoiceStatusDraft
	case stripe.InvoiceStatusUncollectible:
		status = shared_types.InvoiceStatusUncollectible
	case stripe.InvoiceStatusVoid:
		status = shared_types.InvoiceStatusVoid
	default:
		status = shared_types.InvoiceStatusOpen
	}

	invoiceRecord := &shared_types.Invoice{
		BillingAccountID: billingAccountID,
		StripeInvoiceID:  stripeInvoice.ID,
		AmountDue:        int(stripeInvoice.AmountDue),
		AmountPaid:       int(stripeInvoice.AmountPaid),
		Currency:         string(stripeInvoice.Currency),
		Status:           status,
		PeriodStart:      time.Unix(stripeInvoice.PeriodStart, 0),
		PeriodEnd:        time.Unix(stripeInvoice.PeriodEnd, 0),
	}

	if stripeInvoice.HostedInvoiceURL != "" {
		invoiceRecord.InvoiceURL = &stripeInvoice.HostedInvoiceURL
	}
	if stripeInvoice.InvoicePDF != "" {
		invoiceRecord.InvoicePDF = &stripeInvoice.InvoicePDF
	}

	if existingInvoice != nil {
		invoiceRecord.ID = existingInvoice.ID
		return s.storage.UpdateInvoice(invoiceRecord)
	}

	return s.storage.CreateInvoice(invoiceRecord)
}

// handleInvoicePaymentFailed processes invoice.payment_failed events
func (s *BillingService) handleInvoicePaymentFailed(event *stripe.Event) error {
	var stripeInvoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &stripeInvoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Get billing account by customer ID
	billingAccount, err := s.getBillingAccountByCustomerID(stripeInvoice.Customer.ID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get billing account for failed invoice", err.Error())
		return nil // Don't fail, might be a different product
	}

	// Update invoice record with failed status
	if err := s.createOrUpdateInvoiceFromStripe(&stripeInvoice, billingAccount.ID); err != nil {
		s.logger.Log(logger.Error, "Failed to update invoice record for payment failure", err.Error())
		// Continue to update subscription status even if invoice update fails
	}

	// Get subscription and mark as past_due
	if stripeInvoice.Subscription != nil {
		subscription, err := s.storage.GetSubscriptionByStripeID(stripeInvoice.Subscription.ID)
		if err == nil && subscription != nil {
			subscription.Status = shared_types.SubscriptionStatusPastDue
			if err := s.storage.UpdateSubscription(subscription); err != nil {
				s.logger.Log(logger.Error, "Failed to update subscription status to past_due", err.Error())
			}
		}
	}

	return nil
}

// handleSubscriptionUpdated processes customer.subscription.updated events
// Also handles customer.subscription.created events (same logic)
func (s *BillingService) handleSubscriptionUpdated(event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Get billing account by customer ID
	billingAccount, err := s.getBillingAccountByCustomerID(sub.Customer.ID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get billing account for subscription update", err.Error())
		return nil
	}

	// Check if subscription exists
	existingSubscription, _ := s.storage.GetSubscriptionByStripeID(sub.ID)

	// Extract price ID safely (prevent panic if Items.Data is empty)
	var priceID string
	if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		priceID = sub.Items.Data[0].Price.ID
	} else {
		s.logger.Log(logger.Error, "Subscription has no items or price", sub.ID)
		return fmt.Errorf("subscription %s has no items", sub.ID)
	}

	subscription := &shared_types.Subscription{
		BillingAccountID:     billingAccount.ID,
		StripeSubscriptionID: sub.ID,
		StripePriceID:        priceID,
		Status:               mapStripeSubscriptionStatus(sub.Status),
		CurrentPeriodStart:   time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:     time.Unix(sub.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd:    sub.CancelAtPeriodEnd,
	}

	if existingSubscription != nil {
		subscription.ID = existingSubscription.ID
		return s.storage.UpdateSubscription(subscription)
	}

	return s.storage.CreateSubscription(subscription)
}

// handleSubscriptionDeleted processes customer.subscription.deleted events
func (s *BillingService) handleSubscriptionDeleted(event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	subscription, err := s.storage.GetSubscriptionByStripeID(sub.ID)
	if err != nil {
		s.logger.Log(logger.Info, "Subscription not found for deletion", sub.ID)
		return nil
	}

	subscription.Status = shared_types.SubscriptionStatusCanceled
	return s.storage.UpdateSubscription(subscription)
}

// getBillingAccountByCustomerID finds a billing account by Stripe customer ID
func (s *BillingService) getBillingAccountByCustomerID(customerID string) (*shared_types.BillingAccount, error) {
	// This requires a separate query method - for now we'll search through the database
	// In a production system, you'd want to add an index and direct query method
	billingStorage, ok := s.storage.(*storage.BillingStorage)
	if !ok {
		return nil, fmt.Errorf("storage type assertion failed")
	}

	var account shared_types.BillingAccount
	err := billingStorage.DB.NewSelect().
		Model(&account).
		Where("stripe_customer_id = ?", customerID).
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// mapStripeSubscriptionStatus maps Stripe subscription status to our internal status
func mapStripeSubscriptionStatus(status stripe.SubscriptionStatus) shared_types.SubscriptionStatus {
	switch status {
	case stripe.SubscriptionStatusActive:
		return shared_types.SubscriptionStatusActive
	case stripe.SubscriptionStatusPastDue:
		return shared_types.SubscriptionStatusPastDue
	case stripe.SubscriptionStatusCanceled:
		return shared_types.SubscriptionStatusCanceled
	case stripe.SubscriptionStatusIncomplete, stripe.SubscriptionStatusIncompleteExpired:
		return shared_types.SubscriptionStatusIncomplete
	case stripe.SubscriptionStatusTrialing:
		return shared_types.SubscriptionStatusTrialing
	default:
		return shared_types.SubscriptionStatusIncomplete
	}
}
