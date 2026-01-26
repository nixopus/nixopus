package types

import "errors"

// CreateCheckoutRequest represents a request to create a checkout session
type CreateCheckoutRequest struct {
	SuccessURL string `json:"success_url" validate:"required,url"`
	CancelURL  string `json:"cancel_url" validate:"required,url"`
}

// Domain-specific errors
var (
	ErrBillingAccountNotFound     = errors.New("billing account not found")
	ErrSubscriptionNotFound       = errors.New("subscription not found")
	ErrStripeNotConfigured        = errors.New("stripe is not configured")
	ErrInvalidWebhookSignature    = errors.New("invalid webhook signature")
	ErrEventAlreadyProcessed      = errors.New("event already processed")
	ErrPaymentRequired            = errors.New("payment required: deployment limit reached")
	ErrInvalidRequestType         = errors.New("invalid request type")
	ErrMissingSuccessURL          = errors.New("success_url is required")
	ErrMissingCancelURL           = errors.New("cancel_url is required")
	ErrCustomerNotFound           = errors.New("stripe customer not found")
	ErrSubscriptionCreationFailed = errors.New("failed to create subscription")
)

// PaymentRequiredError provides detailed information about why payment is required
type PaymentRequiredError struct {
	DeploymentsUsed  int    `json:"deployments_used"`
	DeploymentsLimit int    `json:"deployments_limit"`
	UpgradeURL       string `json:"upgrade_url,omitempty"`
}

func (e *PaymentRequiredError) Error() string {
	return "payment required: deployment limit reached"
}
