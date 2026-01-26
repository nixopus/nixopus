package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/billing/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type BillingStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

type BillingRepository interface {
	// Billing Account
	GetBillingAccountByOrgID(organizationID uuid.UUID) (*shared_types.BillingAccount, error)
	CreateBillingAccount(account *shared_types.BillingAccount) error
	UpdateBillingAccount(account *shared_types.BillingAccount) error

	// Subscription
	GetActiveSubscription(billingAccountID uuid.UUID) (*shared_types.Subscription, error)
	GetSubscriptionByStripeID(stripeSubscriptionID string) (*shared_types.Subscription, error)
	CreateSubscription(subscription *shared_types.Subscription) error
	UpdateSubscription(subscription *shared_types.Subscription) error

	// Invoice
	GetInvoicesByBillingAccountID(billingAccountID uuid.UUID, limit int) ([]shared_types.Invoice, error)
	GetInvoiceByStripeID(stripeInvoiceID string) (*shared_types.Invoice, error)
	CreateInvoice(invoice *shared_types.Invoice) error
	UpdateInvoice(invoice *shared_types.Invoice) error

	// Payment Events (idempotency)
	IsEventProcessed(stripeEventID string) (bool, error)
	RecordEvent(event *shared_types.PaymentEvent) error

	// Deployment counting
	CountDeployedApplications(organizationID uuid.UUID) (int, error)

	// Transaction support
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) BillingRepository
}

func (s *BillingStorage) BeginTx() (bun.Tx, error) {
	return s.DB.BeginTx(s.Ctx, nil)
}

func (s *BillingStorage) WithTx(tx bun.Tx) BillingRepository {
	return &BillingStorage{
		DB:  s.DB,
		Ctx: s.Ctx,
		tx:  &tx,
	}
}

func (s *BillingStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

// GetBillingAccountByOrgID retrieves a billing account by organization ID
func (s *BillingStorage) GetBillingAccountByOrgID(organizationID uuid.UUID) (*shared_types.BillingAccount, error) {
	var account shared_types.BillingAccount
	err := s.getDB().NewSelect().
		Model(&account).
		Where("organization_id = ?", organizationID).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.ErrBillingAccountNotFound
		}
		return nil, err
	}
	return &account, nil
}

// CreateBillingAccount creates a new billing account
func (s *BillingStorage) CreateBillingAccount(account *shared_types.BillingAccount) error {
	account.ID = uuid.New()
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()
	_, err := s.getDB().NewInsert().Model(account).Exec(s.Ctx)
	return err
}

// UpdateBillingAccount updates an existing billing account
func (s *BillingStorage) UpdateBillingAccount(account *shared_types.BillingAccount) error {
	account.UpdatedAt = time.Now()
	_, err := s.getDB().NewUpdate().
		Model(account).
		WherePK().
		Exec(s.Ctx)
	return err
}

// GetActiveSubscription retrieves the active subscription for a billing account
func (s *BillingStorage) GetActiveSubscription(billingAccountID uuid.UUID) (*shared_types.Subscription, error) {
	var subscription shared_types.Subscription
	err := s.getDB().NewSelect().
		Model(&subscription).
		Where("billing_account_id = ?", billingAccountID).
		Where("status IN (?)", bun.In([]shared_types.SubscriptionStatus{
			shared_types.SubscriptionStatusActive,
			shared_types.SubscriptionStatusTrialing,
		})).
		Order("created_at DESC").
		Limit(1).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}

// GetSubscriptionByStripeID retrieves a subscription by its Stripe ID
func (s *BillingStorage) GetSubscriptionByStripeID(stripeSubscriptionID string) (*shared_types.Subscription, error) {
	var subscription shared_types.Subscription
	err := s.getDB().NewSelect().
		Model(&subscription).
		Where("stripe_subscription_id = ?", stripeSubscriptionID).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.ErrSubscriptionNotFound
		}
		return nil, err
	}
	return &subscription, nil
}

// CreateSubscription creates a new subscription
func (s *BillingStorage) CreateSubscription(subscription *shared_types.Subscription) error {
	subscription.ID = uuid.New()
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = time.Now()
	_, err := s.getDB().NewInsert().Model(subscription).Exec(s.Ctx)
	return err
}

// UpdateSubscription updates an existing subscription
func (s *BillingStorage) UpdateSubscription(subscription *shared_types.Subscription) error {
	subscription.UpdatedAt = time.Now()
	_, err := s.getDB().NewUpdate().
		Model(subscription).
		WherePK().
		Exec(s.Ctx)
	return err
}

// GetInvoicesByBillingAccountID retrieves invoices for a billing account
func (s *BillingStorage) GetInvoicesByBillingAccountID(billingAccountID uuid.UUID, limit int) ([]shared_types.Invoice, error) {
	var invoices []shared_types.Invoice
	query := s.getDB().NewSelect().
		Model(&invoices).
		Where("billing_account_id = ?", billingAccountID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return invoices, nil
}

// GetInvoiceByStripeID retrieves an invoice by its Stripe ID
func (s *BillingStorage) GetInvoiceByStripeID(stripeInvoiceID string) (*shared_types.Invoice, error) {
	var invoice shared_types.Invoice
	err := s.getDB().NewSelect().
		Model(&invoice).
		Where("stripe_invoice_id = ?", stripeInvoiceID).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

// CreateInvoice creates a new invoice
func (s *BillingStorage) CreateInvoice(invoice *shared_types.Invoice) error {
	invoice.ID = uuid.New()
	invoice.CreatedAt = time.Now()
	_, err := s.getDB().NewInsert().Model(invoice).Exec(s.Ctx)
	return err
}

// UpdateInvoice updates an existing invoice
func (s *BillingStorage) UpdateInvoice(invoice *shared_types.Invoice) error {
	_, err := s.getDB().NewUpdate().
		Model(invoice).
		WherePK().
		Exec(s.Ctx)
	return err
}

// IsEventProcessed checks if a Stripe event has already been processed
func (s *BillingStorage) IsEventProcessed(stripeEventID string) (bool, error) {
	exists, err := s.getDB().NewSelect().
		Model((*shared_types.PaymentEvent)(nil)).
		Where("stripe_event_id = ?", stripeEventID).
		Exists(s.Ctx)
	return exists, err
}

// RecordEvent records a processed Stripe event
func (s *BillingStorage) RecordEvent(event *shared_types.PaymentEvent) error {
	event.ID = uuid.New()
	event.ProcessedAt = time.Now()
	_, err := s.getDB().NewInsert().Model(event).Exec(s.Ctx)
	return err
}

// CountDeployedApplications counts the number of deployed applications for an organization
func (s *BillingStorage) CountDeployedApplications(organizationID uuid.UUID) (int, error) {
	count, err := s.getDB().NewSelect().
		Table("applications").
		Join("JOIN application_status ON application_status.application_id = applications.id").
		Where("applications.organization_id = ?", organizationID).
		Where("application_status.status = ?", "deployed").
		Count(s.Ctx)
	return count, err
}
