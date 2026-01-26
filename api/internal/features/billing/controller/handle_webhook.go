package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/billing/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	// MaxBodyBytes limits the size of webhook request body (64KB as per Stripe recommendations)
	MaxBodyBytes = int64(65536)
)

// HandleWebhook processes incoming Stripe webhook events
// This endpoint does not require authentication - it uses Stripe signature verification
// Following Stripe best practices: https://docs.stripe.com/webhooks
func (c *BillingController) HandleWebhook(f fuego.ContextNoBody) (*shared_types.Response, error) {
	r := f.Request()

	// Limit request body size to prevent abuse (Stripe best practice)
	r.Body = http.MaxBytesReader(nil, r.Body, MaxBodyBytes)

	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to read webhook body", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	// Get the Stripe signature from headers
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		c.logger.Log(logger.Error, "Missing Stripe-Signature header", "")
		return nil, fuego.HTTPError{
			Err:    types.ErrInvalidWebhookSignature,
			Status: http.StatusBadRequest,
		}
	}

	// Process the webhook
	if err := c.service.ProcessWebhook(payload, signature); err != nil {
		// Log detailed error for debugging
		c.logger.Log(logger.Error, "Failed to process webhook", fmt.Sprintf("Error: %v, Payload length: %d, Has signature: %t", err, len(payload), signature != ""))

		if err == types.ErrInvalidWebhookSignature {
			c.logger.Log(logger.Error, "Webhook signature verification failed", "Check that STRIPE_WEBHOOK_SECRET matches the secret from 'stripe listen'")
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusBadRequest,
			}
		}

		if err == types.ErrStripeNotConfigured {
			c.logger.Log(logger.Error, "Stripe not configured", "Check that STRIPE_SECRET_KEY and STRIPE_WEBHOOK_SECRET are set")
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusServiceUnavailable,
			}
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Webhook processed successfully",
	}, nil
}
