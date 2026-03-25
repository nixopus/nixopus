package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

var (
	processedDeliveries sync.Map
	deliveryTTL         = 24 * time.Hour
)

func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().Add(-deliveryTTL)
			processedDeliveries.Range(func(key, value any) bool {
				if ts, ok := value.(time.Time); ok && ts.Before(cutoff) {
					processedDeliveries.Delete(key)
				}
				return true
			})
		}
	}()
}

func verifyWebhookSignature(payload []byte, signature, secret string) bool {
	if secret == "" {
		return true
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (c *DeployController) HandleGithubWebhook(f fuego.ContextNoBody) (*types.MessageResponse, error) {
	c.logger.Log(logger.Info, "handling github webhook", "")

	deliveryID := f.Request().Header.Get("X-GitHub-Delivery")
	if deliveryID != "" {
		if _, loaded := processedDeliveries.LoadOrStore(deliveryID, time.Now()); loaded {
			c.logger.Log(logger.Info, "duplicate webhook delivery, skipping", deliveryID)
			return &types.MessageResponse{
				Status:  "success",
				Message: "Duplicate delivery ignored",
			}, nil
		}
	}

	payload, err := io.ReadAll(f.Request().Body)
	if err != nil {
		c.logger.Log(logger.Error, "failed to read webhook payload", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	signature := f.Request().Header.Get("X-Hub-Signature-256")
	if signature == "" {
		c.logger.Log(logger.Error, "missing webhook signature", "")
		return nil, fuego.UnauthorizedError{
			Detail: "missing webhook signature",
			Err:    fmt.Errorf("missing webhook signature"),
		}
	}

	webhookSecret := config.AppConfig.GitHub.WebhookSecret
	if !verifyWebhookSignature(payload, signature, webhookSecret) {
		c.logger.Log(logger.Error, "invalid webhook signature", "")
		return nil, fuego.UnauthorizedError{
			Detail: "invalid webhook signature",
			Err:    fmt.Errorf("invalid webhook signature"),
		}
	}

	eventType := f.Request().Header.Get("X-GitHub-Event")
	if eventType != "push" {
		c.logger.Log(logger.Info, "ignoring non-push event", eventType)
		return &types.MessageResponse{
			Status:  "success",
			Message: "Ignored non-push event",
		}, nil
	}

	var webhookPayload shared_types.WebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		c.logger.Log(logger.Error, "failed to parse webhook payload", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	ref := webhookPayload.Ref
	if !strings.HasPrefix(ref, "refs/heads/") {
		c.logger.Log(logger.Info, "ignoring non-branch ref", ref)
		return &types.MessageResponse{
			Status:  "success",
			Message: fmt.Sprintf("Ignored ref %s", ref),
		}, nil
	}

	err = c.taskService.EnqueueWebhookTask(webhookPayload)
	if err != nil {
		c.logger.Log(logger.Error, "failed to enqueue webhook task", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "github webhook handled successfully", webhookPayload.Repository.FullName)
	return &types.MessageResponse{
		Status:  "success",
		Message: "Github webhook handled successfully",
	}, nil
}
