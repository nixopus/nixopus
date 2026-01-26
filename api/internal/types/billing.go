package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive     SubscriptionStatus = "active"
	SubscriptionStatusPastDue    SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled   SubscriptionStatus = "canceled"
	SubscriptionStatusIncomplete SubscriptionStatus = "incomplete"
	SubscriptionStatusTrialing   SubscriptionStatus = "trialing"
)

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "draft"
	InvoiceStatusOpen          InvoiceStatus = "open"
	InvoiceStatusPaid          InvoiceStatus = "paid"
	InvoiceStatusVoid          InvoiceStatus = "void"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
)

// BillingAccount links organizations to Stripe customers
type BillingAccount struct {
	bun.BaseModel `bun:"table:billing_accounts,alias:ba" swaggerignore:"true"`

	ID                   uuid.UUID `bun:"id,pk,type:uuid" json:"id"`
	OrganizationID       uuid.UUID `bun:"organization_id,notnull,type:uuid" json:"organization_id"`
	StripeCustomerID     *string   `bun:"stripe_customer_id" json:"stripe_customer_id,omitempty"`
	FreeDeploymentsLimit int       `bun:"free_deployments_limit,notnull,default:1" json:"free_deployments_limit"`
	CreatedAt            time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt            time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// Subscription tracks Stripe subscription state
type Subscription struct {
	bun.BaseModel `bun:"table:subscriptions,alias:sub" swaggerignore:"true"`

	ID                   uuid.UUID          `bun:"id,pk,type:uuid" json:"id"`
	BillingAccountID     uuid.UUID          `bun:"billing_account_id,notnull,type:uuid" json:"billing_account_id"`
	StripeSubscriptionID string             `bun:"stripe_subscription_id,notnull" json:"stripe_subscription_id"`
	StripePriceID        string             `bun:"stripe_price_id,notnull" json:"stripe_price_id"`
	Status               SubscriptionStatus `bun:"status,notnull" json:"status"`
	CurrentPeriodStart   time.Time          `bun:"current_period_start,notnull" json:"current_period_start"`
	CurrentPeriodEnd     time.Time          `bun:"current_period_end,notnull" json:"current_period_end"`
	CancelAtPeriodEnd    bool               `bun:"cancel_at_period_end,notnull,default:false" json:"cancel_at_period_end"`
	CreatedAt            time.Time          `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt            time.Time          `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// Invoice stores invoice records from Stripe
type Invoice struct {
	bun.BaseModel `bun:"table:invoices,alias:inv" swaggerignore:"true"`

	ID               uuid.UUID     `bun:"id,pk,type:uuid" json:"id"`
	BillingAccountID uuid.UUID     `bun:"billing_account_id,notnull,type:uuid" json:"billing_account_id"`
	StripeInvoiceID  string        `bun:"stripe_invoice_id,notnull" json:"stripe_invoice_id"`
	AmountDue        int           `bun:"amount_due,notnull" json:"amount_due"`
	AmountPaid       int           `bun:"amount_paid,notnull" json:"amount_paid"`
	Currency         string        `bun:"currency,notnull" json:"currency"`
	Status           InvoiceStatus `bun:"status,notnull" json:"status"`
	InvoiceURL       *string       `bun:"invoice_url" json:"invoice_url,omitempty"`
	InvoicePDF       *string       `bun:"invoice_pdf" json:"invoice_pdf,omitempty"`
	PeriodStart      time.Time     `bun:"period_start,notnull" json:"period_start"`
	PeriodEnd        time.Time     `bun:"period_end,notnull" json:"period_end"`
	CreatedAt        time.Time     `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

// PaymentEvent is the idempotent webhook event log
type PaymentEvent struct {
	bun.BaseModel `bun:"table:payment_events,alias:pe" swaggerignore:"true"`

	ID            uuid.UUID   `bun:"id,pk,type:uuid" json:"id"`
	StripeEventID string      `bun:"stripe_event_id,notnull" json:"stripe_event_id"`
	EventType     string      `bun:"event_type,notnull" json:"event_type"`
	ProcessedAt   time.Time   `bun:"processed_at,notnull,default:current_timestamp" json:"processed_at"`
	Payload       interface{} `bun:"payload,type:jsonb" json:"payload,omitempty"`
}

// BillingStatus represents the billing status response
type BillingStatus struct {
	HasActiveSubscription bool                `json:"has_active_subscription"`
	Subscription          *SubscriptionInfo   `json:"subscription,omitempty"`
	UsageStats            *BillingUsageStats  `json:"usage_stats"`
	BillingAccount        *BillingAccountInfo `json:"billing_account,omitempty"`
}

// SubscriptionInfo contains subscription details for the response
type SubscriptionInfo struct {
	ID                 string             `json:"id"`
	Status             SubscriptionStatus `json:"status"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	CancelAtPeriodEnd  bool               `json:"cancel_at_period_end"`
}

// BillingUsageStats contains deployment usage statistics
type BillingUsageStats struct {
	DeploymentsUsed  int  `json:"deployments_used"`
	DeploymentsLimit int  `json:"deployments_limit"`
	IsUnlimited      bool `json:"is_unlimited"`
}

// BillingAccountInfo contains billing account details for the response
type BillingAccountInfo struct {
	ID                   string `json:"id"`
	FreeDeploymentsLimit int    `json:"free_deployments_limit"`
	HasStripeCustomer    bool   `json:"has_stripe_customer"`
}

// CheckoutSessionResponse represents the response from creating a checkout session
type CheckoutSessionResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
}

// BillingPortalResponse represents the response from creating a billing portal session
type BillingPortalResponse struct {
	PortalURL string `json:"portal_url"`
}

// CanDeployResponse represents the response from the deployment gate check
type CanDeployResponse struct {
	CanDeploy  bool    `json:"can_deploy"`
	Reason     string  `json:"reason,omitempty"`
	UpgradeURL *string `json:"upgrade_url,omitempty"`
}
